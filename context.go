package opencl

/*
#ifdef __APPLE__
#include <OpenCL/opencl.h>
#else
#include <CL/cl.h>
#endif

extern void contextCallback(char *, void *, size_t, void *);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type Context struct {
	context  C.cl_context
	callback func(string, unsafe.Pointer, uintptr, interface{})
	userdata interface{}
	devices  []*Device
}

type ContextProperties struct {
}

//export contextCallback
func contextCallback(errinfo *C.char, private_info unsafe.Pointer, cb C.size_t, user_data unsafe.Pointer) {
	context := *(*Context)(user_data)
	context.callback(C.GoString(errinfo), private_info, uintptr(cb), context.userdata)
}

func CreateContext(properties *ContextProperties, devices []*Device, notify func(string, unsafe.Pointer, uintptr, interface{}), userdata interface{}) (*Context, error) {
	context := Context{nil, notify, userdata, devices}

	ids := make([]C.cl_device_id, len(devices))
	for i, d := range devices {
		ids[i] = d.id
	}

	var callback *[0]byte
	var user unsafe.Pointer
	if notify != nil {
		callback = (*[0]byte)(unsafe.Pointer(C.contextCallback))
		user = unsafe.Pointer(&context)
	}

	var err C.cl_int
	context.context = C.clCreateContext(nil, C.cl_uint(len(ids)), &ids[0], callback, user, &err)
	if err != C.CL_SUCCESS {
		return nil, fmt.Errorf("Error creating context: %d", err)
	}
	return &context, nil
}

func (c Context) Release() error {
	err := C.clReleaseContext(c.context)
	if err != C.CL_SUCCESS {
		return fmt.Errorf("Error releasing context: %d", err)
	}
	return nil
}

func (c Context) Devices() []*Device {
	return c.devices
}

func (c Context) CreateCommandQueue(device Device, properties *CommandQueueProperties) (*CommandQueue, error) {
	return createCommandQueue(c, device, properties)
}

func (c Context) CreateBuffer(flags MemoryFlags, size uintptr, hostPtr unsafe.Pointer) (*Memory, error) {
	return createBuffer(c, flags, size, hostPtr)
}
