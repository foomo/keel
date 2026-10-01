package keel

// ServiceFn creates a [Service]. It is used by [ServiceEnabler] to create a
// fresh service each time it is enabled.
type ServiceFn func() Service
