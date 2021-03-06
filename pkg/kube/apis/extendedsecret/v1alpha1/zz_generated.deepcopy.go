// +build !ignore_autogenerated

/*

Don't alter this file, it was generated.

*/
// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificateRequest) DeepCopyInto(out *CertificateRequest) {
	*out = *in
	if in.AlternativeNames != nil {
		in, out := &in.AlternativeNames, &out.AlternativeNames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.CARef = in.CARef
	out.CAKeyRef = in.CAKeyRef
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificateRequest.
func (in *CertificateRequest) DeepCopy() *CertificateRequest {
	if in == nil {
		return nil
	}
	out := new(CertificateRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtendedSecret) DeepCopyInto(out *ExtendedSecret) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtendedSecret.
func (in *ExtendedSecret) DeepCopy() *ExtendedSecret {
	if in == nil {
		return nil
	}
	out := new(ExtendedSecret)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExtendedSecret) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtendedSecretList) DeepCopyInto(out *ExtendedSecretList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ExtendedSecret, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtendedSecretList.
func (in *ExtendedSecretList) DeepCopy() *ExtendedSecretList {
	if in == nil {
		return nil
	}
	out := new(ExtendedSecretList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ExtendedSecretList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtendedSecretSpec) DeepCopyInto(out *ExtendedSecretSpec) {
	*out = *in
	in.Request.DeepCopyInto(&out.Request)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtendedSecretSpec.
func (in *ExtendedSecretSpec) DeepCopy() *ExtendedSecretSpec {
	if in == nil {
		return nil
	}
	out := new(ExtendedSecretSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtendedSecretStatus) DeepCopyInto(out *ExtendedSecretStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtendedSecretStatus.
func (in *ExtendedSecretStatus) DeepCopy() *ExtendedSecretStatus {
	if in == nil {
		return nil
	}
	out := new(ExtendedSecretStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Request) DeepCopyInto(out *Request) {
	*out = *in
	in.CertificateRequest.DeepCopyInto(&out.CertificateRequest)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Request.
func (in *Request) DeepCopy() *Request {
	if in == nil {
		return nil
	}
	out := new(Request)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretReference) DeepCopyInto(out *SecretReference) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretReference.
func (in *SecretReference) DeepCopy() *SecretReference {
	if in == nil {
		return nil
	}
	out := new(SecretReference)
	in.DeepCopyInto(out)
	return out
}
