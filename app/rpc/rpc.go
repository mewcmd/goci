package rpc

import (
	"fmt"
	"time"
)

//Error is an error type suitable for sending over an rpc response
type Error string

//Error implemented the error interface for the Error type
func (r Error) Error() string {
	return string(r)
}

//Errorf is a helper method for creating an rpc error that can be transmitted
//in json.
func Errorf(format string, v ...interface{}) error {
	return Error(fmt.Sprintf(format, v...))
}

//Wrap is a convenience function to wrap whatever error we get in an rpc Error
func Wrap(errp *error) {
	if errp == nil {
		return
	}
	err := *errp
	if _, ok := err.(Error); err != nil && ok {
		*errp = Errorf("%s", err)
	}
}

//AnnounceArgs is the argument type of the Announce function
type AnnounceArgs struct {
	GOOS, GOARCH string //the goos/goarch of the service
	Type         string //either "Builder" or "Runner"
	URL          string //the url of the service to make rpc calls
}

//AnnounceReply is the reply type of the Announce function
type AnnounceReply struct {
	//Key is the datastore key that corresponds to the service if successful
	Key string
}

//RemoveArgs is the argument type of the Remove function
type RemoveArgs struct {
	//Key is the datastore key that corresponds to the service to be removed
	Key  string
	Kind string
}

//None is an empty rpc element
type None struct{}

//Work is an incoming work item to generate the builds for a given revision and
//import path. If Revision is empty, the revision chosen by go get is used. If
//Subpackages is true, it will build binaries for all subpackages of the import
//path as well.
type Work struct {
	Revision    string
	ImportPath  string
	Subpackages bool

	//VCSHint is an optional parameter that specifies the version control system
	//used by the package. If set to the empty string, we will search for the
	//system by looking for the metadata directory.
	VCSHint string
}

//Distill makes a Work able to be sent in to the queue.
func (w Work) Distill() (Work, string) {
	return w, ""
}

//BuilderTask is a task sent to a Builder
type BuilderTask struct {
	Work     Work   //the Work item to be completed
	Key      string //the datastore key for the Work item (forward to runner)
	ID       string //the id of the test (forward to runner)
	WorkRev  int    //the revision of the work item (forward to runner)
	Runner   string //the rpc url of the runner for this task
	Response string //the rpc url of the response (forward to the runner)
}

//RunnerTask is a task sent by a Builder to a runner
type RunnerTask struct {
	Key        string    //the key of the work item to pass into response
	ID         string    //the ID of the test
	WorkRev    int       //the revision of the work item
	Revision   string    //the revision we ended up testing to pass into response
	RevDate    time.Time //the time this revision was made to pass into response
	Tests      []RunTest //the set of binarys to be executed
	WontBuilds []Output  //the set of tests that failed to build
	Response   string    //the rpc url of the response
}

//RunTest represents an individual binary to be installed and run.
type RunTest struct {
	BinaryURL  string //the url to download the binary
	SourceURL  string //the url to download the tarball
	ImportPath string //the import path of the packge the binary is testing
	Config     Config //the configuration for this test
}

//RunnerResponse is the response from the Runner to the tracker
type RunnerResponse struct {
	Key      string    //the key for the Work item
	ID       string    //the ID of the test
	WorkRev  int       //the revision of the work item
	Revision string    //the revision we ended up testing
	RevDate  time.Time //the time this revision was made
	Tests    []Output  //the list of tests
}

//BuilderResponse is the response from the Builder if the build failed for any
//reason.
type BuilderResponse struct {
	Key      string    //the key of the work item
	ID       string    //the ID of the test
	WorkRev  int       //the expected revision of work item
	Error    string    //the error in setting up the builds
	Revision string    //the revision of the work item (if known)
	RevDate  time.Time //the time the revision was commit (if known)
}

//DispatchResponse is the response from the dispatcher to the response handler
//saying that it is unable to get a successful response from the work item and
//it has failed too many times.
type DispatchResponse struct {
	Key     string //the key of the work item
	Error   string //the error in the dispatch
	WorkRev int    //the revision of the work document
}

//Output is a type that wraps the output of a build, be it the actual output or
//the error produced.
type Output struct {
	ImportPath string     //the import path of the binary that produced the output
	Config     Config     //the configuration for the test
	Type       OutputType //the type of output (Success/WontBuild/Error)
	Output     string     //the output of the test
}

//OutputType is an enumeration of types of outputs.
type OutputType string

const (
	OutputSuccess   OutputType = "Success"
	OutputWontBuild OutputType = "WontBuild"
	OutputError     OutputType = "Error"
)

//TestResponse is the args type for the Post method on a Runner.
type TestResponse struct {
	ID     string //the ID for the test
	Output Output //the output of the test
}

//TestRequest is the args type for the Request methdo on the Runner.
type TestRequest struct {
	ID    string //the ID of the test
	Index int    //the index of the test to be run
}
