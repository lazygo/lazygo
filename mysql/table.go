package mysql

import "fmt"

type Table string

func (t Table) String() string {
	return buildK(string(t))
}

func (t Table) IsValid() bool {
	return t != ""
}

func (t Table) As(alias string) Table {
	return Table(t.String() + " AS " + buildK(alias))
}

func (t Table) LeftJoin(table Table, a, b string) Table {
	return Table(fmt.Sprintf("%s LEFT JOIN %s ON %s = %s", t.String(), table, buildK(a), buildK(b)))
}
