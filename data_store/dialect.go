package data_store

// type dialect interface {
// 	Select(query *QueryImpl) string
// }

// type DefaultDialect struct {
// }

// func (self *DefaultDialect) Select(query *QueryImpl) string {

// 	a := fmt.Sprintf("SELECT %v FROM %v", query.column, query.tableName)
// 	if query.joinStr != "" {
// 		a = fmt.Sprintf("%v %v", a, query.joinStr)
// 	}
// 	if query.where != "" {
// 		a = fmt.Sprintf("%v WHERE %v", a, query.where)
// 	}
// 	if query.groupBy != "" {
// 		a = fmt.Sprintf("%v %v", a, query.groupBy)
// 	}
// 	if query.having != "" {
// 		a = fmt.Sprintf("%v %v", a, query.having)
// 	}
// 	if query.order != "" {
// 		a = query.Sprintf("%v ORDER BY %v", a, query.order)
// 	}
// 	if orm.offset > 0 {
// 		a = query.Sprintf("%v LIMIT %v, %v", a, query.offset, query.limit)
// 	} else if orm.limit > 0 {
// 		a = query.Sprintf("%v LIMIT %v", a, query.limit)
// 	}
// }

// func (orm *QueryImpl) generateSql() (a string) {
//     if orm.ParamIdentifier == "mssql" {
//         if orm.offset > 0 {
//             a = fmt.Sprintf("select ROW_NUMBER() OVER(order by %v )as rownum,%v from %v",
//                 orm.primaryKey,
//                 orm.column,
//                 orm.tableName)
//             if orm.where != "" {
//                 a = fmt.Sprintf("%v WHERE %v", a, orm.where)
//             }
//             a = fmt.Sprintf("select * from (%v) "+
//                 "as a where rownum between %v and %v",
//                 a,
//                 orm.offset,
//                 orm.limit)
//         } else if orm.limit > 0 {
//             a = fmt.Sprintf("SELECT top %v %v FROM %v",
//                 orm.limit, orm.column, orm.tableName)
//             if orm.where != "" {
//                 a = fmt.Sprintf("%v WHERE %v", a, orm.where)
//             }
//             if orm.groupBy != "" {
//                 a = fmt.Sprintf("%v %v", a, orm.groupBy)
//             }
//             if orm.having != "" {
//                 a = fmt.Sprintf("%v %v", a, orm.having)
//             }
//             if orm.order != "" {
//                 a = fmt.Sprintf("%v ORDER BY %v", a, orm.order)
//             }
//         } else {
//             a = fmt.Sprintf("SELECT %v FROM %v", orm.column, orm.tableName)
//             if orm.where != "" {
//                 a = fmt.Sprintf("%v WHERE %v", a, orm.where)
//             }
//             if orm.groupBy != "" {
//                 a = fmt.Sprintf("%v %v", a, orm.groupBy)
//             }
//             if orm.having != "" {
//                 a = fmt.Sprintf("%v %v", a, orm.having)
//             }
//             if orm.order != "" {
//                 a = fmt.Sprintf("%v ORDER BY %v", a, orm.order)
//             }
//         }
//     } else {
//         a = fmt.Sprintf("SELECT %v FROM %v", orm.column, orm.tableName)
//         if orm.joinStr != "" {
//             a = fmt.Sprintf("%v %v", a, orm.joinStr)
//         }
//         if orm.where != "" {
//             a = fmt.Sprintf("%v WHERE %v", a, orm.where)
//         }
//         if orm.groupBy != "" {
//             a = fmt.Sprintf("%v %v", a, orm.groupBy)
//         }
//         if orm.having != "" {
//             a = fmt.Sprintf("%v %v", a, orm.having)
//         }
//         if orm.order != "" {
//             a = fmt.Sprintf("%v ORDER BY %v", a, orm.order)
//         }
//         if orm.offset > 0 {
//             a = fmt.Sprintf("%v LIMIT %v, %v", a, orm.offset, orm.limit)
//         } else if orm.limit > 0 {
//             a = fmt.Sprintf("%v LIMIT %v", a, orm.limit)
//         }
//     }
//     return
// }
