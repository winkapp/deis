package portunes

import (
	"container/list"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Masterminds/cookoo"
	"github.com/Masterminds/cookoo/safely"
	"github.com/Masterminds/cookoo/web"
)

var DefaultHistoryLen = 100

// Battery describes a series of exams, together with configuration.
type Battery struct {
	// Exams lists the exams to run.
	Exams []*Exam `yaml:"exams"`
	// Notifiers lists the things that can be notified about exam results.
	Notifiers []*Notifier `yaml:"notifiers"`
	// HistoryLen indicates the maximum size of the history for any given exam.
	HistoryLen int `yaml:"historyLen"`
}

// Exam describes a named test and its operational parameters.
//
// Why call it Exam instead of Test? Because Go ascribes special meaning to
// things named Test.
type Exam struct {
	// Name is the name of the exam that this will call.
	Name string `yaml:"name"`
	// Interval is the string representation of how frequently this exam is run.
	//
	// This supports an duration string handled by Go's time.ParseDuration
	//
	// Examples:
	//
	// 	10s: run every ten seconds.
	// 	3h10m: run every three hours and ten minutes.
	Interval string `yaml:"interval"`

	// Depends lists the other tests that this depends on.
	//
	// This test will only be executed if the other groups have already passed
	// there most recent test.
	Depends []string `yaml:"depends"`

	// Notify lists the groups that should be notified.
	Notify []string `yaml:"notify"`
}

// Duration returns a time.Duration representation of the Interval.
func (e *Exam) Duration() (time.Duration, error) {
	return time.ParseDuration(e.Interval)
}

// Notifier describes a service or thing that can be notified of an Exam's outcome.
type Notifier struct {
	Name   string
	Config map[string]string
}

// Portunes is the test-running service.
type Portunes struct {
	battery  *Battery
	Registry *cookoo.Registry
	router   *cookoo.Router
	cxt      cookoo.Context
	results  *results
}

func New(b *Battery) *Portunes {

	if b.HistoryLen == 0 {
		b.HistoryLen = DefaultHistoryLen
	}

	reg, router, c := cookoo.Cookoo()
	p := &Portunes{
		battery:  b,
		Registry: reg,
		router:   router,
		cxt:      cookoo.SyncContext(c),
		results:  newResults(b.HistoryLen),
	}

	return p
}

// Run starts the Portunes battery of tests, but does not start the API server.
func (p *Portunes) Run(cancel chan bool) error {

	if err := p.preflight(); err != nil {
		return err
	}

	for _, exam := range p.battery.Exams {
		exam := exam
		safely.Go(func() { p.runExam(exam, cancel) })
	}

	return nil
}

func (p *Portunes) runExam(exam *Exam, cancel chan bool) {

	// Add an initial record that we have not done anything yet.
	ir := &Result{
		Exam:    exam.Name,
		Date:    time.Now(),
		Outcome: Unknown,
		Message: "Exam loaded, but not run yet.",
	}
	p.results.Add(ir)

	dur, err := exam.Duration()
	if err != nil {
		// XXX: We could instead give this a default duration. With the
		// current method, the test will never run, and will hence always show
		// up in the Unknown state. We could also panic (rightly), since the
		// Duration should have been checked already by preflight.
		return
	}
	ticker := time.NewTicker(dur)
	for {
		select {
		case now := <-ticker.C:
			c := p.cxt.Copy()
			res := &Result{Exam: exam.Name, Date: now}
			println("Running test", exam.Name)
			if err := p.router.HandleRequest(exam.Name, c, true); err != nil {
				res.Outcome = Fail
				res.Message = err.Error()
			} else {
				res.Outcome = Pass
			}
			p.results.Add(res)
		case <-cancel:
			ticker.Stop()
			return
		}
	}
}

// preflight runs a pre-flight check to see if the battery is actually runnable.
func (p *Portunes) preflight() error {
	for _, e := range p.battery.Exams {
		if !p.router.HasRoute(e.Name) {
			return fmt.Errorf("no exam named '%s' found", e.Name)
		}

		for _, d := range e.Depends {
			if !p.router.HasRoute(d) {
				return fmt.Errorf("no dependency named '%s' found", d)
			}
		}

		if _, err := e.Duration(); err != nil {
			return fmt.Errorf("interval '%s' is not a valid duration: %s", e.Interval, err)
		}

		for _, n := range e.Notify {
			found := false
			for _, ns := range p.battery.Notifiers {
				if ns.Name == n {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("notifier '%s' not found", n)
			}
		}
	}

	return nil
}

// ServerHTTP starts an API HTTP server on the given address.
func (p *Portunes) ServeHTTP(addr string) {
	reg, router, c := cookoo.Cookoo()

	c.Put("results", p.results)

	reg.Route("@encode", "Encode JSON and write to HTTP").
		Does(jsonify, "json").Using("data").From("cxt:data").
		Does(web.Flush, "_").Using("content").From("cxt:json").Using("contentType").WithDefault("application/json")

	reg.Route("GET /healthz", "Health check endpoint").Does(pong, "pong")

	reg.Route("GET /v1/battery", "List all exams in battery").
		Does(getBattery, "data").Using("results").From("cxt:results").
		Includes("@encode")

	reg.Route("GET /v1/exam/*", "Get the status of an exam").
		Does(getExam, "data").
		Using("results").From("cxt:results").
		Using("exam").From("path:2").
		Includes("@encode")

	reg.Route("GET /v1/exam/*/history", "Get the history for an exam").
		Does(getExamHistory, "data").
		Using("results").From("cxt:results").
		Using("exam").From("path:2").
		Includes("@encode")

	h := web.NewCookooHandler(reg, router, c)
	http.ListenAndServe(addr, h)
}

const (
	// Pass indicates an exam passed
	Pass Outcome = "pass"
	// Fail indicates an exam failed
	Fail Outcome = "fail"
	// Unknown indicates that exam is in an unknown state, and has possibly never
	// been run."
	Unknown Outcome = "unknown"
)

// Outcome describes an exam outcome.
type Outcome string

// Result describes the outcome of a particular exam run.
type Result struct {
	// Exam is the name of the exam.
	Exam string
	// Date is the time this exam was run.
	Date time.Time
	// Outcome is the result of thsi exam.
	Outcome Outcome
	// Message contains any optional text information about the exam, such as
	// why it failed.
	Message string
}

// ResultStorer describes the ability to store exam results.
type ResultStorer interface {
	// Add adds a result to result storage.
	Add(*Result)
	// List lists the results for a particular exam.
	List(string) []*Result
	// Len lists the history length for a particular exam.
	Len(string) int
	// Empty removes an exam's history.
	Empty(string)
	// Last gets the last result for a given exam. This will be nil if no
	// results have been stored.
	Last(string) *Result

	// Exams returns a map of all known exams and their most recent outcome.
	Exams() map[string]string
}

type results struct {
	mx  sync.RWMutex
	res map[string]*list.List
	max int
}

func newResults(max int) *results {
	return &results{res: map[string]*list.List{}, max: max}
}

func (r *results) Add(res *Result) {
	r.mx.Lock()
	defer r.mx.Unlock()

	l, ok := r.res[res.Exam]
	if !ok {
		l = list.New()
		r.res[res.Exam] = l
	}

	l.PushFront(res)
	if l.Len() > r.max {
		l.Remove(l.Back())
	}
}

func (r *results) Empty(name string) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if _, ok := r.res[name]; ok {
		delete(r.res, name)
	}
}

func (r *results) Last(name string) *Result {
	r.mx.RLock()
	defer r.mx.RUnlock()
	l, ok := r.res[name]
	if !ok {
		return nil
	}

	first := l.Front()
	if first.Value == nil {
		return nil
	}

	return first.Value.(*Result)
}

func (r *results) List(name string) []*Result {
	r.mx.RLock()
	defer r.mx.RUnlock()
	l, ok := r.res[name]
	if !ok {
		return []*Result{}
	}
	rs := make([]*Result, l.Len())
	i := 0
	for e := l.Front(); e != nil; e = e.Next() {
		rs[i] = e.Value.(*Result)
		i++
	}

	return rs
}

func (r *results) Len(name string) int {
	r.mx.RLock()
	defer r.mx.RUnlock()
	l, ok := r.res[name]
	if !ok {
		return 0
	}
	return l.Len()
}

func (r *results) Exams() map[string]string {
	res := make(map[string]string, len(r.res))
	for k, v := range r.res {
		outcome := Unknown
		if v != nil && v.Front() != nil && v.Front().Value != nil {
			outcome = v.Front().Value.(*Result).Outcome
		}
		res[k] = string(outcome)
	}
	return res
}

// pong answers a healthz.
func pong(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	w := c.Get("http.ResponseWriter", nil).(http.ResponseWriter)
	w.Write([]byte("Pong"))
	w.WriteHeader(http.StatusOK)
	return true, nil
}

// jsonify encodes a thing into JSON
//
// Params:
// - data (interface{})
//
// Returns:
// - []byte of JSON data
func jsonify(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	data := p.Get("data", "").(interface{})
	return json.Marshal(data)
}

// getBattery returns a battery of tests.
func getBattery(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	results := p.Get("results", &results{}).(ResultStorer)
	return results.Exams(), nil
}

// getExam gets the current status for a single exam.
func getExam(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	results := p.Get("results", &results{}).(ResultStorer)
	exam := p.Get("exam", "").(string)
	return results.Last(exam), nil
}

// getExamHistory gets the history for an exam.
func getExamHistory(c cookoo.Context, p *cookoo.Params) (interface{}, cookoo.Interrupt) {
	results := p.Get("results", &results{}).(ResultStorer)
	exam := p.Get("exam", "").(string)
	res := results.List(exam)
	return res, nil
}
