package bear

type Segment struct {
	Template string
	Values   []interface{}
}

func New(template string, values []interface{}) *Segment {
	return &Segment{Template: template, Values: values}
}

func (s Segment) Tuple() (string, []interface{}) {
	return s.Template, s.Values
}
