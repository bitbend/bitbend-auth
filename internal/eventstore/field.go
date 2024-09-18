package eventstore

type FieldValue struct {
	Value     any
	IsIndexed bool
	IsUnique  bool
}

type Field struct {
	Aggregate  *Aggregate
	FieldName  string
	FieldValue FieldValue
}
