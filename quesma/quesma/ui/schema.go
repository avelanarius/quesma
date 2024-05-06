package ui

import (
	"fmt"
	"mitmproxy/quesma/clickhouse"
	"mitmproxy/quesma/util"
	"sort"
	"strings"
)

func (qmc *QuesmaManagementConsole) generateSchema() []byte {
	type menuEntry struct {
		label  string
		target string
	}

	var menuEntries []menuEntry

	type tableColumn struct {
		name             string
		typeName         string
		isAttribute      bool
		isFullTextSearch bool
		warning          *string
	}

	buffer := newBufferWithHead()
	buffer.Write(generateTopNavigation("schema"))
	buffer.Html(`<main id="schema">`)

	if qmc.logManager != nil {

		// Not sure if we should read directly from the TableMap or we should use the Snapshot of it.
		// Let's leave it as is for now.
		schema := qmc.logManager.GetTableDefinitions()

		tableNames := schema.Keys()
		sort.Strings(tableNames)

		buffer.Html("\n<table>")

		for i, tableName := range tableNames {
			table, ok := schema.Load(tableName)
			if !ok {
				continue
			}

			id := fmt.Sprintf("schema-table-%d", i)
			var menu menuEntry
			menu.label = table.Name
			menu.target = fmt.Sprintf("#%s", id)
			menuEntries = append(menuEntries, menu)

			buffer.Html(`<tr class="tableName"`)
			buffer.Html(fmt.Sprintf(` id="%s"`, id))
			buffer.Html(`>`)
			buffer.Html(`<th colspan=2><h2>`)
			buffer.Html(`Table: `)
			buffer.Text(table.Name)

			if table.Comment != "" {
				buffer.Text(" (")
				buffer.Text(table.Comment)
				buffer.Text(")")
			}

			buffer.Html(`</h2></th>`)
			buffer.Html(`</tr>`)

			buffer.Html(`<tr>`)
			buffer.Html(`<th>`)
			buffer.Html(`Name`)
			buffer.Html(`</th>`)
			buffer.Html(`<th>`)
			buffer.Html(`Type`)
			buffer.Html(`</th>`)
			buffer.Html(`</tr>`)

			var columnNames []string
			var columnMap = make(map[string]tableColumn)

			// standard columns, visible for the user
			for k := range table.Cols {
				c := tableColumn{}

				c.name = k
				if table.Cols[k].Type != nil {
					c.typeName = table.Cols[k].Type.StringWithNullable()
				} else {
					c.typeName = "n/a"
				}

				c.isAttribute = false
				c.isFullTextSearch = table.Cols[k].IsFullTextMatch

				columnNames = append(columnNames, k)
				columnMap[k] = c
			}

			for _, a := range qmc.config.AliasFields(table.Name) {

				// check for collisions
				if field, collide := columnMap[a.SourceFieldName]; collide {
					field.warning = util.Pointer("alias declared with the same name")
					columnMap[a.SourceFieldName] = field
					continue
				}

				// check if target exists
				c := tableColumn{}
				c.name = a.SourceFieldName
				if aliasedField, ok := columnMap[a.TargetFieldName]; ok {
					c.typeName = fmt.Sprintf("alias of '%s', %s", a.TargetFieldName, aliasedField.typeName)
					c.isFullTextSearch = aliasedField.isFullTextSearch
					c.isAttribute = aliasedField.isAttribute
				} else {
					c.warning = util.Pointer("alias points to non-existing field '" + a.TargetFieldName + "'")
					c.typeName = "dangling alias"
				}

				columnNames = append(columnNames, a.SourceFieldName)
				columnMap[a.SourceFieldName] = c
			}

			// columns added by Quesma, not visible for the user
			//
			// this part is based on addOurFieldsToCreateTableQuery in log_manager.go
			attributes := table.Config.GetAttributes()
			if len(attributes) > 0 {
				for _, a := range attributes {
					_, ok := table.Cols[a.KeysArrayName]
					if !ok {
						c := tableColumn{}
						c.name = a.KeysArrayName
						c.typeName = clickhouse.CompoundType{Name: "Array", BaseType: clickhouse.NewBaseType("String")}.StringWithNullable()
						c.isAttribute = true
						columnNames = append(columnNames, c.name)
						columnMap[c.name] = c
					}
					_, ok = table.Cols[a.ValuesArrayName]
					if !ok {
						c := tableColumn{}
						c.name = a.ValuesArrayName
						c.typeName = clickhouse.CompoundType{Name: "Array", BaseType: a.Type}.StringWithNullable()
						c.isAttribute = true
						columnNames = append(columnNames, c.name)
						columnMap[c.name] = c
					}
				}
			}

			sort.Strings(columnNames)

			for _, columnName := range columnNames {
				column, ok := columnMap[columnName]
				if !ok {
					continue
				}

				buffer.Html(`<tr class="`)

				if column.isAttribute {
					buffer.Html(`columnAttribute `)
				}
				if column.warning != nil {
					buffer.Html(`columnWarning `)
				}
				buffer.Html(`column`)

				buffer.Html(`">`)
				buffer.Html(`<td class="columnName">`)

				buffer.Text(column.name)
				buffer.Html(`</td>`)
				buffer.Html(`<td class="columnType">`)

				buffer.Text(column.typeName)
				if column.isFullTextSearch {
					buffer.Html(` <i>(Full text match)</i>`)
				}

				if column.warning != nil {
					buffer.Html(` <span class="columnWarningText">WARNING: `)
					buffer.Text(*column.warning)
					buffer.Html(`</span>`)
				}

				buffer.Html(`</td>`)
				buffer.Html(`</tr>`)
			}

		}

		buffer.Html("\n</table>")

	} else {
		buffer.Html(`<p>Schema is not available</p>`)
	}

	buffer.Html("\n<table>")
	buffer.Html(`<tr class="tableName" id="quesma-config">`)
	buffer.Html(`<th colspan=3><h2>`)
	buffer.Html(`Quesma Config`)
	buffer.Html(`</h2></th>`)
	buffer.Html(`</tr>`)

	buffer.Html(`<tr>`)
	buffer.Html(`<th>`)
	buffer.Html(`Name Pattern`)
	buffer.Html(`</th>`)
	buffer.Html(`<th>`)
	buffer.Html(`Enabled?`)
	buffer.Html(`</th>`)
	buffer.Html(`<th>`)
	buffer.Html(`Full Text Search Fields`)
	buffer.Html(`</th>`)

	buffer.Html(`</tr>`)

	for _, cfg := range qmc.config.IndexConfig {
		buffer.Html(`<tr>`)
		buffer.Html(`<td>`)
		buffer.Text(cfg.Name)
		buffer.Html(`</td>`)
		buffer.Html(`<td>`)
		if cfg.Enabled {
			buffer.Text("true")
		} else {
			buffer.Text("false")
		}
		buffer.Html(`</td>`)

		buffer.Html(`<td>`)
		buffer.Text(strings.Join(cfg.FullTextFields, ", "))
		buffer.Html(`</td>`)

		buffer.Html(`</tr>`)
	}

	buffer.Html("\n</table>")

	buffer.Html("\n</main>\n\n")

	buffer.Html(`<div class="menu">`)
	buffer.Html("\n<h2>Menu</h2>")

	buffer.Html(`<h3>Admin actions</h3>`)
	buffer.Html(`<ul>`)

	buffer.Html(`<li><button hx-post="/schema/reload" hx-target="body">Reload Schemas</button></li>`)

	buffer.Html(`</ul>`)

	buffer.Html(`<h3>Tables:</h3>`)

	buffer.Html("<ol>")

	for _, menu := range menuEntries {
		buffer.Html(`<li><a href="`)
		buffer.Text(menu.target)
		buffer.Html(`">`)
		buffer.Text(menu.label)
		buffer.Html(`</a></li>`)
	}

	buffer.Html("</ol>")

	buffer.Html(`<h3><a href="#quesma-config">Jump to Quesma Config</a></h3>`)

	buffer.Html("\n</div>")

	buffer.Html("\n</body>")
	buffer.Html("\n</html>")
	return buffer.Bytes()
}