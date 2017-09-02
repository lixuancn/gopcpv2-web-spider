package scheduler

import (
	"gopcpv2-web-spider/module"
	"gopcpv2-web-spider/toolkit/buffer"
	"testing"
	"gopcpv2-web-spider/errors"
)

func TestErrorGen(t *testing.T) {
	simpleErrMsg := "testing error"
	expectedErrType := errors.ERROR_TYPE_SCHEDULER
	err := genError(simpleErrMsg)
	ce, ok := err.(errors.CrawlerError)
	if !ok {
		t.Fatalf("Inconsistent error type: expected: %T, actual: %T",
			errors.NewCrawlerError("", ""), err)
	}
	if ce.Type() != expectedErrType {
		t.Fatalf("Inconsistent error type string: expected: %q, actual: %q",
			expectedErrType, ce.Type())
	}
	expectedErrMsg := "Crawler Error: 调度器失败: " + simpleErrMsg
	if ce.Error() != expectedErrMsg {
		t.Fatalf("Inconsistent error message: expected: %q, actual: %q",
			expectedErrMsg, ce.Error())
	}
}

func TestErrorGenByError(t *testing.T) {
	simpleErrMsg := "testing error"
	simpleErr := errors.NewCrawlerError(errors.ERROR_TYPE_SCHEDULER, simpleErrMsg)
	expectedErrType := errors.ERROR_TYPE_SCHEDULER
	err := genErrorByError(simpleErr)
	ce, ok := err.(errors.CrawlerError)
	if !ok {
		t.Fatalf("Inconsistent error type: expected: %T, actual: %T",
			errors.NewCrawlerError("", ""), err)
	}
	if ce.Type() != expectedErrType {
		t.Fatalf("Inconsistent error type string: expected: %q, actual: %q",
			expectedErrType, ce.Type())
	}
}

func TestParameterErrorGen(t *testing.T) {
	simpleErrMsg := "testing error"
	expectedErrType := errors.ERROR_TYPE_SCHEDULER
	err := genParameterError(simpleErrMsg)
	ce, ok := err.(errors.CrawlerError)
	if !ok {
		t.Fatalf("Inconsistent error type: expected: %T, actual: %T",
			errors.NewCrawlerError("", ""), err)
	}
	if ce.Type() != expectedErrType {
		t.Fatalf("Inconsistent error type string: expected: %q, actual: %q",
			expectedErrType, ce.Type())
	}
	expectedErrMsg := "Crawler Error: 调度器失败: illegal parameter: " + simpleErrMsg
	if ce.Error() != expectedErrMsg {
		t.Fatalf("Inconsistent error message: expected: %q, actual: %q",
			expectedErrMsg, ce.Error())
	}
}

func TestErrorSend(t *testing.T) {
	cerr := errors.NewCrawlerError(
		errors.ERROR_TYPE_SCHEDULER, "testing error")
	mid := module.MID("")
	buffer, _ := buffer.NewPool(10, 2)
	if !sendError(cerr, mid, buffer) {
		t.Fatalf("Couldn't send error! (error: %s, MID: %s, buffer: %#v)",
			cerr, mid, buffer)
	}
	err := errors.NewCrawlerError(errors.ERROR_TYPE_SCHEDULER, "testing error")
	if !sendError(err, mid, buffer) {
		t.Fatalf("Couldn't send error! (error: %s, MID: %s, buffer: %#v)",
			err, mid, buffer)
	}
	mids := []module.MID{
		module.MID("D0"),
		module.MID("A0"),
		module.MID("P0"),
	}
	for _, mid := range mids {
		if !sendError(err, mid, buffer) {
			t.Fatalf("Couldn't send error! (error: %s, MID: %s, buffer: %#v)",
				err, mid, buffer)
		}
	}
	if sendError(nil, mid, buffer) {
		t.Fatalf("It still can send error with nil error!")
	}
	if sendError(err, mid, nil) {
		t.Fatalf("It still can send error with nil buffer!")
	}
	buffer.Close()
	if sendError(err, mid, buffer) {
		t.Fatalf("It still can send error with closed buffer!")
	}
}