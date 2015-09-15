package portunes

import (
	"container/list"
	"errors"
	"testing"
	"time"

	"github.com/Masterminds/cookoo"
)

func TestPreflight(t *testing.T) {
	b := basicBattery()
	p := New(b)
	if err := p.preflight(); err == nil {
		t.Error("Should have failed preflight check.")
	}

	p.Registry.AddRoutes(
		cookoo.Route{Name: "test1"},
		cookoo.Route{Name: "test2"},
		cookoo.Route{Name: "test3"},
	)

	if err := p.preflight(); err != nil {
		t.Errorf("Failed preflight check: %s", err)
	}
}

func TestResults(t *testing.T) {
	var rss ResultStorer = &results{max: 3, res: map[string]*list.List{}}

	res1 := &Result{Exam: "test1", Date: time.Now(), Outcome: Pass}
	rss.Add(res1)
	if rss.Len("test1") != 1 {
		t.Errorf("Expected len 1, got %d", rss.Len("test1"))
	}

	rss.Empty("test1")
	if rss.Len("test1") != 0 {
		t.Errorf("Expected empty list, got %d", rss.Len("test1"))
	}

	for i := 0; i < 4; i++ {
		rss.Add(&Result{Exam: "test1", Date: time.Now()})
	}
	if rss.Len("test1") != 3 {
		t.Errorf("Expected len 3, got %d", rss.Len("test1"))
	}
}

func TestRunExam(t *testing.T) {
	p := New(basicBattery())
	basicRoutes(p.Registry)
	cancel := make(chan bool)

	go p.runExam(p.battery.Exams[0], cancel)
	time.Sleep(7 * time.Millisecond)
	cancel <- true

	if p.results.Len("test1") != 2 {
		t.Fatalf("Expected two entries in the log, got %d", p.results.Len("test1"))
	}

	res := p.results.List("test1")
	if res[0].Outcome != Pass {
		t.Errorf("Expected latest test to be marked 'pass'. Got '%s'", res[0].Outcome)
	}
	if res[1].Outcome != Unknown {
		t.Errorf("Expected first test to be marked 'unknown'. Got '%s'", res[1].Outcome)
	}

	// Test a failed exam.
	go p.runExam(p.battery.Exams[2], cancel)
	time.Sleep(13 * time.Millisecond)
	cancel <- true

	if p.results.Len("test3") != 2 {
		t.Fatalf("Expected two entries in the log, got %d", p.results.Len("test3"))
	}

	res = p.results.List("test3")
	if res[0].Outcome != Fail {
		t.Errorf("Expected latest test to be marked 'fail'. Got '%s'", res[0].Outcome)
	}
	if res[1].Outcome != Unknown {
		t.Errorf("Expected first test to be marked 'unknown'. Got '%s'", res[1].Outcome)
	}
}

func basicBattery() *Battery {
	b := &Battery{
		Exams: []*Exam{
			&Exam{Name: "test1", Interval: "5ms"},
			&Exam{Name: "test2", Interval: "1s", Depends: []string{"test1"}},
			&Exam{Name: "test3", Interval: "10ms", Notify: []string{"group1"}},
		},
		Notifiers: []*Notifier{
			&Notifier{Name: "group1"},
		},
	}
	return b
}
func basicRoutes(reg *cookoo.Registry) {
	reg.AddRoutes(
		cookoo.Route{
			Name: "test1",
			Does: []cookoo.Task{
				&cookoo.Cmd{
					Name: "test1",
					Fn:   basicCommand,
				},
			},
		},
		cookoo.Route{
			Name: "test2",
			Does: []cookoo.Task{
				cookoo.Cmd{
					Name: "test2",
					Fn:   basicCommand,
				},
			},
		},
		cookoo.Route{
			Name: "test3",
			Does: []cookoo.Task{
				cookoo.Cmd{
					Name: "test3",
					Fn:   failedCommand,
				},
			},
		},
	)
}

// basicCommand puts 'true' into the context.
// Returns:
// 	bool true
func basicCommand(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	return true, nil
}

// failedCommand always returns an error.
func failedCommand(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	return nil, errors.New("I'm a creep. I'm a weirdo.")
}
