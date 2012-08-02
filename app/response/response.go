// +build !goci

//package response provides handlers for storing responses
package response

import (
	"appengine"
	"appengine/datastore"
	gorcp "code.google.com/p/gorilla/rpc"
	gojson "code.google.com/p/gorilla/rpc/json"
	"httputil"
	"net/http"
	"queue"
	"rpc"
	"strings"
	"time"
)

func init() {
	//create a new rpc server
	s := gorcp.NewServer()
	s.RegisterCodec(gojson.NewCodec(), "application/json")

	//add the response queue
	s.RegisterService(Response{}, "")

	http.Handle("/response", s)
}

//Response is a service that records Runner responses
type Response struct{}

//Post is the rpc method that the Runner uses to give a response about an item.
func (Response) Post(req *http.Request, args *rpc.RunnerResponse, resp *rpc.None) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	//create our context
	ctx := appengine.NewContext(req)
	ctx.Infof("Storing runner result")
	ctx.Infof("%+v", args)

	//make sure the TaskInfo for this request still exists, and if so, remove it
	found, err := queue.FindTaskInfo(ctx, args.ID)

	if err != nil {
		ctx.Errorf("error finding task info: %v", err)
		return
	}

	//if we didn't find our TaskInfo, just bail
	if !found {
		ctx.Errorf("Got a late response")
		return
	}

	//get the key of the work item
	key := httputil.FromString(args.Key)

	//create a WorkResult
	w := &WorkResult{
		Success: true,
		When:    time.Now(),
	}

	//store it
	wkey := datastore.NewIncompleteKey(ctx, "WorkResult", key)
	if wkey, err = datastore.Put(ctx, wkey, w); err != nil {
		ctx.Errorf("Error storing WorkResult: %v", err)
		return
	}

	//store the test results
	for _, out := range args.Outputs {
		//make a TestResult
		t := &TestResult{
			ImportPath: out.ImportPath,
			Revision:   args.Revision,
			RevDate:    args.RevDate,
			When:       time.Now(),
			Output:     out.Output,
			Passed:     strings.HasSuffix(out.Output, "\nPASS\n"),
		}

		//store it
		tkey := datastore.NewIncompleteKey(ctx, "TestResult", wkey)
		if _, err = datastore.Put(ctx, tkey, t); err != nil {
			ctx.Errorf("Error storing TestResult: %v", err)
			return
		}
	}

	//we did it!
	return
}

//Error is used when there were any errors in building the test
func (Response) Error(req *http.Request, args *rpc.BuilderResponse, resp *rpc.None) (err error) {
	//wrap our error on the way out
	defer rpc.Wrap(&err)

	//create the context
	ctx := appengine.NewContext(req)
	ctx.Infof("Storing a builder error")
	ctx.Infof("%+v", args)

	//make sure the TaskInfo for this request still exists, and if so, remove it
	found, err := queue.FindTaskInfo(ctx, args.ID)

	//make sure we check the error
	if err != nil {
		ctx.Errorf("error finding task info: %v", err)
		return
	}

	//if we didn't find it, just bail
	if !found {
		ctx.Errorf("Got a late response")
		return
	}

	//get the key of the work item
	key := httputil.FromString(args.Key)

	//create a WorkResult
	w := &WorkResult{
		Success: false,
		When:    time.Now(),
		Error:   args.Error,
	}

	//store it
	wkey := datastore.NewIncompleteKey(ctx, "WorkResult", key)
	if wkey, err = datastore.Put(ctx, wkey, w); err != nil {
		ctx.Errorf("Error storing WorkResult: %v", err)
		return
	}

	//store all the build errors
	for _, out := range args.BuildErrors {
		//create a BuildFailure
		b := &BuildFailure{
			ImportPath: out.ImportPath,
			Revision:   args.Revision,
			RevDate:    args.RevDate,
			When:       time.Now(),
			Output:     out.Error,
		}

		//store it
		bkey := datastore.NewIncompleteKey(ctx, "BuildFailure", wkey)
		if _, err = datastore.Put(ctx, bkey, b); err != nil {
			ctx.Errorf("Error storing BuildFailure: %v", err)
			return
		}
	}

	//we did it!
	return
}
