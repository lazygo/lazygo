package mysql

import "fmt"

type Table string

func (t Table) String() string {
	table := string(t)
	if isSimple(table) {
		return buildK(table)
	}
	return table
}

func (t Table) IsValid() bool {
	return t != ""
}

func (t Table) As(alias string) Table {
	return Table(t.String() + " AS " + alias)
}

func (t Table) LeftJoin(table Table, a, b string) Table {
	return Table(fmt.Sprintf("%s LEFT JOIN %s ON %s = %s", t.String(), table, buildK(a), buildK(b)))
}
