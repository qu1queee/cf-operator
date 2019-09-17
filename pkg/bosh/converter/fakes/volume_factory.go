// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"code.cloudfoundry.org/cf-operator/pkg/bosh/bpm"
	"code.cloudfoundry.org/cf-operator/pkg/bosh/converter"
	"code.cloudfoundry.org/cf-operator/pkg/bosh/disk"
	"code.cloudfoundry.org/cf-operator/pkg/bosh/manifest"
)

type FakeVolumeFactory struct {
	GenerateBPMDisksStub        func(string, *manifest.InstanceGroup, bpm.Configs, string) (disk.BPMResourceDisks, error)
	generateBPMDisksMutex       sync.RWMutex
	generateBPMDisksArgsForCall []struct {
		arg1 string
		arg2 *manifest.InstanceGroup
		arg3 bpm.Configs
		arg4 string
	}
	generateBPMDisksReturns struct {
		result1 disk.BPMResourceDisks
		result2 error
	}
	generateBPMDisksReturnsOnCall map[int]struct {
		result1 disk.BPMResourceDisks
		result2 error
	}
	GenerateDefaultDisksStub        func(string, string, string, string) disk.BPMResourceDisks
	generateDefaultDisksMutex       sync.RWMutex
	generateDefaultDisksArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 string
	}
	generateDefaultDisksReturns struct {
		result1 disk.BPMResourceDisks
	}
	generateDefaultDisksReturnsOnCall map[int]struct {
		result1 disk.BPMResourceDisks
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeVolumeFactory) GenerateBPMDisks(arg1 string, arg2 *manifest.InstanceGroup, arg3 bpm.Configs, arg4 string) (disk.BPMResourceDisks, error) {
	fake.generateBPMDisksMutex.Lock()
	ret, specificReturn := fake.generateBPMDisksReturnsOnCall[len(fake.generateBPMDisksArgsForCall)]
	fake.generateBPMDisksArgsForCall = append(fake.generateBPMDisksArgsForCall, struct {
		arg1 string
		arg2 *manifest.InstanceGroup
		arg3 bpm.Configs
		arg4 string
	}{arg1, arg2, arg3, arg4})
	fake.recordInvocation("GenerateBPMDisks", []interface{}{arg1, arg2, arg3, arg4})
	fake.generateBPMDisksMutex.Unlock()
	if fake.GenerateBPMDisksStub != nil {
		return fake.GenerateBPMDisksStub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.generateBPMDisksReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeVolumeFactory) GenerateBPMDisksCallCount() int {
	fake.generateBPMDisksMutex.RLock()
	defer fake.generateBPMDisksMutex.RUnlock()
	return len(fake.generateBPMDisksArgsForCall)
}

func (fake *FakeVolumeFactory) GenerateBPMDisksCalls(stub func(string, *manifest.InstanceGroup, bpm.Configs, string) (disk.BPMResourceDisks, error)) {
	fake.generateBPMDisksMutex.Lock()
	defer fake.generateBPMDisksMutex.Unlock()
	fake.GenerateBPMDisksStub = stub
}

func (fake *FakeVolumeFactory) GenerateBPMDisksArgsForCall(i int) (string, *manifest.InstanceGroup, bpm.Configs, string) {
	fake.generateBPMDisksMutex.RLock()
	defer fake.generateBPMDisksMutex.RUnlock()
	argsForCall := fake.generateBPMDisksArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeVolumeFactory) GenerateBPMDisksReturns(result1 disk.BPMResourceDisks, result2 error) {
	fake.generateBPMDisksMutex.Lock()
	defer fake.generateBPMDisksMutex.Unlock()
	fake.GenerateBPMDisksStub = nil
	fake.generateBPMDisksReturns = struct {
		result1 disk.BPMResourceDisks
		result2 error
	}{result1, result2}
}

func (fake *FakeVolumeFactory) GenerateBPMDisksReturnsOnCall(i int, result1 disk.BPMResourceDisks, result2 error) {
	fake.generateBPMDisksMutex.Lock()
	defer fake.generateBPMDisksMutex.Unlock()
	fake.GenerateBPMDisksStub = nil
	if fake.generateBPMDisksReturnsOnCall == nil {
		fake.generateBPMDisksReturnsOnCall = make(map[int]struct {
			result1 disk.BPMResourceDisks
			result2 error
		})
	}
	fake.generateBPMDisksReturnsOnCall[i] = struct {
		result1 disk.BPMResourceDisks
		result2 error
	}{result1, result2}
}

func (fake *FakeVolumeFactory) GenerateDefaultDisks(arg1 string, arg2 string, arg3 string, arg4 string) disk.BPMResourceDisks {
	fake.generateDefaultDisksMutex.Lock()
	ret, specificReturn := fake.generateDefaultDisksReturnsOnCall[len(fake.generateDefaultDisksArgsForCall)]
	fake.generateDefaultDisksArgsForCall = append(fake.generateDefaultDisksArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 string
		arg4 string
	}{arg1, arg2, arg3, arg4})
	fake.recordInvocation("GenerateDefaultDisks", []interface{}{arg1, arg2, arg3, arg4})
	fake.generateDefaultDisksMutex.Unlock()
	if fake.GenerateDefaultDisksStub != nil {
		return fake.GenerateDefaultDisksStub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.generateDefaultDisksReturns
	return fakeReturns.result1
}

func (fake *FakeVolumeFactory) GenerateDefaultDisksCallCount() int {
	fake.generateDefaultDisksMutex.RLock()
	defer fake.generateDefaultDisksMutex.RUnlock()
	return len(fake.generateDefaultDisksArgsForCall)
}

func (fake *FakeVolumeFactory) GenerateDefaultDisksCalls(stub func(string, string, string, string) disk.BPMResourceDisks) {
	fake.generateDefaultDisksMutex.Lock()
	defer fake.generateDefaultDisksMutex.Unlock()
	fake.GenerateDefaultDisksStub = stub
}

func (fake *FakeVolumeFactory) GenerateDefaultDisksArgsForCall(i int) (string, string, string, string) {
	fake.generateDefaultDisksMutex.RLock()
	defer fake.generateDefaultDisksMutex.RUnlock()
	argsForCall := fake.generateDefaultDisksArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeVolumeFactory) GenerateDefaultDisksReturns(result1 disk.BPMResourceDisks) {
	fake.generateDefaultDisksMutex.Lock()
	defer fake.generateDefaultDisksMutex.Unlock()
	fake.GenerateDefaultDisksStub = nil
	fake.generateDefaultDisksReturns = struct {
		result1 disk.BPMResourceDisks
	}{result1}
}

func (fake *FakeVolumeFactory) GenerateDefaultDisksReturnsOnCall(i int, result1 disk.BPMResourceDisks) {
	fake.generateDefaultDisksMutex.Lock()
	defer fake.generateDefaultDisksMutex.Unlock()
	fake.GenerateDefaultDisksStub = nil
	if fake.generateDefaultDisksReturnsOnCall == nil {
		fake.generateDefaultDisksReturnsOnCall = make(map[int]struct {
			result1 disk.BPMResourceDisks
		})
	}
	fake.generateDefaultDisksReturnsOnCall[i] = struct {
		result1 disk.BPMResourceDisks
	}{result1}
}

func (fake *FakeVolumeFactory) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.generateBPMDisksMutex.RLock()
	defer fake.generateBPMDisksMutex.RUnlock()
	fake.generateDefaultDisksMutex.RLock()
	defer fake.generateDefaultDisksMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeVolumeFactory) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ converter.VolumeFactory = new(FakeVolumeFactory)
