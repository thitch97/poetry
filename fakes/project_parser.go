package fakes

import "sync"

type ProjectParser struct {
	ParseCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Detected bool
			Version  string
			Err      error
		}
		Stub func(string) (bool, string, error)
	}
}

func (f *ProjectParser) Parse(param1 string) (bool, string, error) {
	f.ParseCall.Lock()
	defer f.ParseCall.Unlock()
	f.ParseCall.CallCount++
	f.ParseCall.Receives.Path = param1
	if f.ParseCall.Stub != nil {
		return f.ParseCall.Stub(param1)
	}
	return f.ParseCall.Returns.Detected, f.ParseCall.Returns.Version, f.ParseCall.Returns.Err
}
