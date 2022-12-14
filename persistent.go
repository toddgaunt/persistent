// Package persistent provides data structures for Go that that always preserve
// the previous version of themselves when any operation is performed on them.
// Since these structures are effectively immutable, this allows for liberal
// memory sharing between instances of these structures as operations are
// perfomed, rather than simplistic memory copying and duplication.
package persistent
