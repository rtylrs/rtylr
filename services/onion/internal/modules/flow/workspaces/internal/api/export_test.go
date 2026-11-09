package api

// RecheckRealtimeAccess lets the external tests record the access rechecks a
// handler asks for, instead of publishing them to a running hub.
var RecheckRealtimeAccess = &recheckRealtimeAccess
