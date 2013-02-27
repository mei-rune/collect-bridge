package mdb

import (
	"testing"
)

func createPrintedBook() map[string]interface{} {
	return map[string]interface{}{"type": "book",
		"name":       "ac reference",
		"publish_at": 1,
		"author":     "mfk",
		"isbn":       "sfasfd-sf-ssf",
		"page_count": 300}
}
func createMagazine() map[string]interface{} {
	return map[string]interface{}{"type": "magazine",
		"name":       "reader",
		"publish_at": 1,
		"journal_id": 4,
		"page_count": 300}
}
func createHelpFile(website_id string) map[string]interface{} {
	return map[string]interface{}{"type": "help_file",
		"name":              "msdn",
		"publish_at":        1,
		"compressed_format": "cab",
		"bytes":             3 * 1024 * 1024,
		"website_id":        website_id}
}
func createOnlineTutorial(website_id string) map[string]interface{} {
	return map[string]interface{}{"type": "online_tutorial",
		"name":       "msdn",
		"publish_at": 1,
		"bytes":      3 * 1024 * 1024,
		"website_id": website_id}
}

func createWebsite() map[string]interface{} {
	return map[string]interface{}{"url": "www.mdb.org"}
}

func saveIt(t *testing.T, target string, attributes map[string]interface{}) string {
	return create(t, target, attributes)
}
func change(attributes, merge map[string]interface{}) map[string]interface{} {
	for k, v := range merge {
		attributes[k] = v
	}
	return attributes
}

func checkModelCount(t *testing.T, target string, params map[string]string, all int) {
	if c, err := count(target, params); all != c {
		t.Errorf("%d != len(%s), actual is %d, %v", all, target, c, err)
	}
	res := findBy(t, target, params)
	if all != len(res) {
		t.Errorf("%d != len(%s), actual is %d", all, target, len(res))
	}
}
func clearAll(t *testing.T) {
	deleteById(t, "document", "all")
	deleteById(t, "zip_file", "all")
	deleteById(t, "website", "all")
	deleteById(t, "topic", "all")
	deleteById(t, "printer", "all")
}
func TestFindAllDocuments(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 8; i++ {
		saveIt(t, "document", createPrintedBook())
		saveIt(t, "document", createHelpFile(website1))
		saveIt(t, "document", createOnlineTutorial(website2))
	}
	checkModelCount(t, "book", map[string]string{}, 8)
	checkModelCount(t, "magazine", map[string]string{}, 0)
	checkModelCount(t, "printed_document", map[string]string{}, 8)

	checkModelCount(t, "help_file", map[string]string{}, 8)
	checkModelCount(t, "online_tutorial", map[string]string{}, 8)
	checkModelCount(t, "online_document", map[string]string{}, 16)

	checkModelCount(t, "document", map[string]string{}, 24)
}

func TestFindDocuments(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
	}
	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
	}

	checkModelCount(t, "book", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "magazine", map[string]string{"name": "cc"}, 0)
	checkModelCount(t, "printed_document", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "help_file", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "online_document", map[string]string{"name": "cc"}, 8)

	checkModelCount(t, "document", map[string]string{"name": "cc"}, 12)
}

func TestDeleteAllPrintedBooks(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 8; i++ {
		saveIt(t, "document", createPrintedBook())
		saveIt(t, "document", createMagazine())
		saveIt(t, "document", createHelpFile(website1))
		saveIt(t, "document", createOnlineTutorial(website2))
	}

	deleteById(t, "book", "all")

	checkModelCount(t, "book", map[string]string{}, 0)
	checkModelCount(t, "magazine", map[string]string{}, 8)
	checkModelCount(t, "printed_document", map[string]string{}, 8)

	checkModelCount(t, "help_file", map[string]string{}, 8)
	checkModelCount(t, "online_tutorial", map[string]string{}, 8)
	checkModelCount(t, "online_document", map[string]string{}, 16)
	checkModelCount(t, "document", map[string]string{}, 24)
}

func TestDeleteAllPrintedDocuments(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 8; i++ {
		saveIt(t, "document", createPrintedBook())
		saveIt(t, "document", createMagazine())
		saveIt(t, "document", createHelpFile(website1))
		saveIt(t, "document", createOnlineTutorial(website2))
	}

	deleteById(t, "printed_document", "all")

	checkModelCount(t, "book", map[string]string{}, 0)
	checkModelCount(t, "magazine", map[string]string{}, 0)
	checkModelCount(t, "printed_document", map[string]string{}, 0)

	checkModelCount(t, "help_file", map[string]string{}, 8)
	checkModelCount(t, "online_tutorial", map[string]string{}, 8)
	checkModelCount(t, "online_document", map[string]string{}, 16)

	checkModelCount(t, "document", map[string]string{}, 16)
}

func TestDeleteAllDocuments(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 8; i++ {
		saveIt(t, "document", createPrintedBook())
		saveIt(t, "document", createMagazine())
		saveIt(t, "document", createHelpFile(website1))
		saveIt(t, "document", createOnlineTutorial(website2))
	}

	deleteById(t, "document", "all")

	checkModelCount(t, "book", map[string]string{}, 0)
	checkModelCount(t, "magazine", map[string]string{}, 0)
	checkModelCount(t, "printed_document", map[string]string{}, 0)

	checkModelCount(t, "help_file", map[string]string{}, 0)
	checkModelCount(t, "online_tutorial", map[string]string{}, 0)
	checkModelCount(t, "online_document", map[string]string{}, 0)
	checkModelCount(t, "document", map[string]string{}, 0)

}

func TestPrintedBooksDeleteWithQuery(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
	}
	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
	}

	deleteBy(t, "book", map[string]string{"name": "aa"})

	checkModelCount(t, "book", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "magazine", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "printed_document", map[string]string{"name": "aa"}, 4)

	checkModelCount(t, "book", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "magazine", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "printed_document", map[string]string{"name": "cc"}, 8)

	checkModelCount(t, "help_file", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "online_document", map[string]string{"name": "aa"}, 8)

	checkModelCount(t, "document", map[string]string{"name": "cc"}, 16)

	checkModelCount(t, "help_file", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "online_document", map[string]string{"name": "cc"}, 8)

	checkModelCount(t, "document", map[string]string{"name": "aa"}, 12)
	checkModelCount(t, "document", map[string]string{"name": "cc"}, 16)
}

func TestPrintedDocumentsDeleteWithQuery(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
	}
	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
	}

	deleteBy(t, "printed_document", map[string]string{"name": "aa"})

	checkModelCount(t, "book", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "magazine", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "printed_document", map[string]string{"name": "aa"}, 0)

	checkModelCount(t, "book", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "magazine", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "printed_document", map[string]string{"name": "cc"}, 8)

	checkModelCount(t, "help_file", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "online_document", map[string]string{"name": "aa"}, 8)

	checkModelCount(t, "help_file", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "online_document", map[string]string{"name": "cc"}, 8)

	checkModelCount(t, "document", map[string]string{"name": "aa"}, 8)
	checkModelCount(t, "document", map[string]string{"name": "cc"}, 16)
}

func TestDocumentsDeleteWithQuery(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
	}
	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
	}

	deleteBy(t, "document", map[string]string{"name": "aa"})

	checkModelCount(t, "book", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "magazine", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "printed_document", map[string]string{"name": "aa"}, 0)

	checkModelCount(t, "help_file", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "online_document", map[string]string{"name": "aa"}, 0)

	checkModelCount(t, "document", map[string]string{"name": "aa"}, 0)

	checkModelCount(t, "book", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "magazine", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "printed_document", map[string]string{"name": "cc"}, 8)

	checkModelCount(t, "help_file", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "cc"}, 4)
	checkModelCount(t, "online_document", map[string]string{"name": "cc"}, 8)

	checkModelCount(t, "document", map[string]string{"name": "cc"}, 16)
}

func TestUpdateAllDocuments(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
	}
	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
	}
	updateBy(t, "book", map[string]string{}, map[string]interface{}{"name": "bb"})

	checkModelCount(t, "book", map[string]string{"name": "bb"}, 8)
	checkModelCount(t, "book", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "book", map[string]string{"name": "cc"}, 0)

	checkModelCount(t, "magazine", map[string]string{"name": "bb"}, 0)
	checkModelCount(t, "magazine", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "magazine", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "help_file", map[string]string{"name": "bb"}, 0)
	checkModelCount(t, "help_file", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "help_file", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "online_tutorial", map[string]string{"name": "bb"}, 0)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "cc"}, 4)
}

func TestBaseClassUpdateAllDocuments(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())
	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
	}
	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
	}

	updateBy(t, "printed_document", map[string]string{}, map[string]interface{}{"name": "bb"})

	checkModelCount(t, "book", map[string]string{"name": "bb"}, 8)
	checkModelCount(t, "book", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "book", map[string]string{"name": "cc"}, 0)

	checkModelCount(t, "magazine", map[string]string{"name": "bb"}, 8)
	checkModelCount(t, "magazine", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "magazine", map[string]string{"name": "cc"}, 0)

	checkModelCount(t, "help_file", map[string]string{"name": "bb"}, 0)
	checkModelCount(t, "help_file", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "help_file", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "online_tutorial", map[string]string{"name": "bb"}, 0)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "cc"}, 4)
}

func TestTopClassUpdateAllDocuments(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
	}
	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
	}

	updateBy(t, "document", map[string]string{}, map[string]interface{}{"name": "bb"})

	checkModelCount(t, "book", map[string]string{"name": "bb"}, 8)
	checkModelCount(t, "book", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "book", map[string]string{"name": "cc"}, 0)

	checkModelCount(t, "magazine", map[string]string{"name": "bb"}, 8)
	checkModelCount(t, "magazine", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "magazine", map[string]string{"name": "cc"}, 0)

	checkModelCount(t, "help_file", map[string]string{"name": "bb"}, 8)
	checkModelCount(t, "help_file", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "help_file", map[string]string{"name": "cc"}, 0)

	checkModelCount(t, "online_tutorial", map[string]string{"name": "bb"}, 8)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "cc"}, 0)
}

func TestUpdateDocuments(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
	}
	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
	}
	updateBy(t, "book", map[string]string{"name": "bb"}, map[string]interface{}{"name": "aa"})

	checkModelCount(t, "book", map[string]string{"name": "bb"}, 4)
	checkModelCount(t, "book", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "book", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "magazine", map[string]string{"name": "bb"}, 0)
	checkModelCount(t, "magazine", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "magazine", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "help_file", map[string]string{"name": "bb"}, 0)
	checkModelCount(t, "help_file", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "help_file", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "online_tutorial", map[string]string{"name": "bb"}, 0)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "cc"}, 4)
}

func TestBaseClassUpdateDocuments(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
	}
	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
	}

	updateBy(t, "printed_document", map[string]string{"name": "bb"}, map[string]interface{}{"name": "aa"})

	checkModelCount(t, "book", map[string]string{"name": "bb"}, 4)
	checkModelCount(t, "book", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "book", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "magazine", map[string]string{"name": "bb"}, 4)
	checkModelCount(t, "magazine", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "magazine", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "help_file", map[string]string{"name": "bb"}, 0)
	checkModelCount(t, "help_file", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "help_file", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "online_tutorial", map[string]string{"name": "bb"}, 0)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "aa"}, 4)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "cc"}, 4)
}

func TestTopClassUpdateDocuments(t *testing.T) {
	clearAll(t)
	website1 := saveIt(t, "website", createWebsite())
	website2 := saveIt(t, "website", createWebsite())
	//website3 := saveIt(t, "website", createWebsite())

	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
	}
	for i := 0; i < 4; i++ {
		saveIt(t, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
		saveIt(t, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
	}

	updateBy(t, "document", map[string]string{"name": "bb"}, map[string]interface{}{"name": "aa"})
	//     Document.update("name = ?", "name= ?", "bb", "aa");

	checkModelCount(t, "book", map[string]string{"name": "bb"}, 4)
	checkModelCount(t, "book", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "book", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "magazine", map[string]string{"name": "bb"}, 4)
	checkModelCount(t, "magazine", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "magazine", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "help_file", map[string]string{"name": "bb"}, 4)
	checkModelCount(t, "help_file", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "help_file", map[string]string{"name": "cc"}, 4)

	checkModelCount(t, "online_tutorial", map[string]string{"name": "bb"}, 4)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "aa"}, 0)
	checkModelCount(t, "online_tutorial", map[string]string{"name": "cc"}, 4)
}
