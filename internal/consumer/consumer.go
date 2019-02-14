// Package consumer models the usage manifests that named downstream services
// publish to declare which parts of a contract they actually depend on.
//
// The impact analyzer joins the set of detected contract changes against these
// manifests so that a report can say "change X breaks consumer Y" rather than
// merely "change X exists". A manifest is deliberately coarse: it lists the
// endpoints a consumer calls and the fields it reads or writes. That is enough
// to compute a precise blast radius without forcing consumers to publish their
// entire source tree.
package consumer
