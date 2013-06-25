package ds

import (
	"commons/types"
	"testing"
)

/**
 * STI Case:
 * <ul>
 * <li>Magazine       extends PrintedDocument extends Document{documents}
 * <li>PrintedBook    extends PrintedDocument extends Document{documents}
 * <li>OnlineTutorial extends OnlineDocument  extends Document{documents}
 * <li>HelpFile       extends OnlineDocument  extends Document{documents}
 * </ul>
 * Related associations:
 * <ul>
 * <li>Document       --has many--   ZipFile
 * <li>OnlineDocument --belongs to-- Website
 * <li>HelpFile       --has many--   Topic
 * <li>PrintedBook    --belongs to-- Printer
 * </ul>
 */
func createDocument() map[string]interface{} {
	return map[string]interface{}{"type": "document",
		"name":       "doc_test",
		"publish_at": 1}
}
func createPrintedDocument() map[string]interface{} {
	return map[string]interface{}{"type": "printed_document",
		"name":       "doc_test",
		"publish_at": 1,
		"page_count": 300}
}

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

func saveIt(t *testing.T, client *Client, target string, attributes map[string]interface{}) string {
	return create(t, client, target, attributes)
}
func change(attributes, merge map[string]interface{}) map[string]interface{} {
	for k, v := range merge {
		attributes[k] = v
	}
	return attributes
}

func checkModelCount(t *testing.T, client *Client, target string, params map[string]string, all int64) {
	if c := count(t, client, target, params); all != c {
		t.Errorf("%d != len(%s), actual is %d", all, target, c)
	}
	res := findBy(t, client, target, params)
	if all != int64(len(res)) {
		t.Errorf("%d != len(%s), actual is %d", all, target, len(res))
	}
}
func clearAll(t *testing.T, client *Client) {
	deleteBy(t, client, "document", map[string]string{})
	deleteBy(t, client, "zip_file", map[string]string{})
	deleteBy(t, client, "website", map[string]string{})
	deleteBy(t, client, "topic", map[string]string{})
	deleteBy(t, client, "printer", map[string]string{})
}
func TestSTIByFindAllDocuments(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 8; i++ {
			saveIt(t, client, "document", createDocument())
			saveIt(t, client, "document", createPrintedDocument())
			saveIt(t, client, "document", createPrintedBook())
			saveIt(t, client, "document", createHelpFile(website1))
			saveIt(t, client, "document", createOnlineTutorial(website2))

		}
		checkModelCount(t, client, "book", map[string]string{}, 8)
		checkModelCount(t, client, "magazine", map[string]string{}, 0)
		checkModelCount(t, client, "printed_document", map[string]string{}, 16)

		checkModelCount(t, client, "help_file", map[string]string{}, 8)
		checkModelCount(t, client, "online_tutorial", map[string]string{}, 8)
		checkModelCount(t, client, "online_document", map[string]string{}, 16)

		checkModelCount(t, client, "document", map[string]string{}, 40)
	})
}

func TestSTIByFindDocuments(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
		}
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
		}

		checkModelCount(t, client, "book", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "magazine", map[string]string{"name": "cc"}, 0)
		checkModelCount(t, client, "printed_document", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "help_file", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "online_document", map[string]string{"name": "cc"}, 8)

		checkModelCount(t, client, "document", map[string]string{"name": "cc"}, 12)
	})
}

func TestSTIByDeleteAllPrintedBooks(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 8; i++ {
			saveIt(t, client, "document", createPrintedBook())
			saveIt(t, client, "document", createMagazine())
			saveIt(t, client, "document", createHelpFile(website1))
			saveIt(t, client, "document", createOnlineTutorial(website2))
		}

		deleteBy(t, client, "book", map[string]string{})

		checkModelCount(t, client, "book", map[string]string{}, 0)
		checkModelCount(t, client, "magazine", map[string]string{}, 8)
		checkModelCount(t, client, "printed_document", map[string]string{}, 8)

		checkModelCount(t, client, "help_file", map[string]string{}, 8)
		checkModelCount(t, client, "online_tutorial", map[string]string{}, 8)
		checkModelCount(t, client, "online_document", map[string]string{}, 16)
		checkModelCount(t, client, "document", map[string]string{}, 24)
	})
}

func TestSTIByDeleteAllPrintedDocuments(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 8; i++ {
			saveIt(t, client, "document", createPrintedBook())
			saveIt(t, client, "document", createMagazine())
			saveIt(t, client, "document", createHelpFile(website1))
			saveIt(t, client, "document", createOnlineTutorial(website2))
		}

		deleteBy(t, client, "printed_document", map[string]string{})

		checkModelCount(t, client, "book", map[string]string{}, 0)
		checkModelCount(t, client, "magazine", map[string]string{}, 0)
		checkModelCount(t, client, "printed_document", map[string]string{}, 0)

		checkModelCount(t, client, "help_file", map[string]string{}, 8)
		checkModelCount(t, client, "online_tutorial", map[string]string{}, 8)
		checkModelCount(t, client, "online_document", map[string]string{}, 16)

		checkModelCount(t, client, "document", map[string]string{}, 16)
	})
}

func TestSTIByDeleteAllPrintedDocuments2(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 8; i++ {
			saveIt(t, client, "document", createDocument())
			saveIt(t, client, "document", createPrintedDocument())
			saveIt(t, client, "document", createPrintedBook())
			saveIt(t, client, "document", createMagazine())
			saveIt(t, client, "document", createHelpFile(website1))
			saveIt(t, client, "document", createOnlineTutorial(website2))
		}

		deleteBy(t, client, "printed_document", map[string]string{})

		checkModelCount(t, client, "book", map[string]string{}, 0)
		checkModelCount(t, client, "magazine", map[string]string{}, 0)
		checkModelCount(t, client, "printed_document", map[string]string{}, 0)

		checkModelCount(t, client, "help_file", map[string]string{}, 8)
		checkModelCount(t, client, "online_tutorial", map[string]string{}, 8)
		checkModelCount(t, client, "online_document", map[string]string{}, 16)

		checkModelCount(t, client, "document", map[string]string{}, 24)
	})
}

func TestSTIByDeleteAllDocuments(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 8; i++ {
			saveIt(t, client, "document", createPrintedBook())
			saveIt(t, client, "document", createMagazine())
			saveIt(t, client, "document", createHelpFile(website1))
			saveIt(t, client, "document", createOnlineTutorial(website2))
		}

		deleteBy(t, client, "document", map[string]string{})

		checkModelCount(t, client, "book", map[string]string{}, 0)
		checkModelCount(t, client, "magazine", map[string]string{}, 0)
		checkModelCount(t, client, "printed_document", map[string]string{}, 0)

		checkModelCount(t, client, "help_file", map[string]string{}, 0)
		checkModelCount(t, client, "online_tutorial", map[string]string{}, 0)
		checkModelCount(t, client, "online_document", map[string]string{}, 0)
		checkModelCount(t, client, "document", map[string]string{}, 0)
	})
}

func TestSTIByPrintedBooksDeleteWithQuery(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
		}
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
		}

		deleteBy(t, client, "book", map[string]string{"name": "aa"})

		checkModelCount(t, client, "book", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "magazine", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "printed_document", map[string]string{"name": "aa"}, 4)

		checkModelCount(t, client, "book", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "magazine", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "printed_document", map[string]string{"name": "cc"}, 8)

		checkModelCount(t, client, "help_file", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "online_document", map[string]string{"name": "aa"}, 8)

		checkModelCount(t, client, "document", map[string]string{"name": "cc"}, 16)

		checkModelCount(t, client, "help_file", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "online_document", map[string]string{"name": "cc"}, 8)

		checkModelCount(t, client, "document", map[string]string{"name": "aa"}, 12)
		checkModelCount(t, client, "document", map[string]string{"name": "cc"}, 16)
	})
}

func TestSTIByPrintedDocumentsDeleteWithQuery(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
		}
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
		}

		deleteBy(t, client, "printed_document", map[string]string{"name": "aa"})

		checkModelCount(t, client, "book", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "magazine", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "printed_document", map[string]string{"name": "aa"}, 0)

		checkModelCount(t, client, "book", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "magazine", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "printed_document", map[string]string{"name": "cc"}, 8)

		checkModelCount(t, client, "help_file", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "online_document", map[string]string{"name": "aa"}, 8)

		checkModelCount(t, client, "help_file", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "online_document", map[string]string{"name": "cc"}, 8)

		checkModelCount(t, client, "document", map[string]string{"name": "aa"}, 8)
		checkModelCount(t, client, "document", map[string]string{"name": "cc"}, 16)
	})
}

func TestSTIByDocumentsDeleteWithQuery(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
		}
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "bb"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
		}

		deleteBy(t, client, "document", map[string]string{"name": "aa"})

		checkModelCount(t, client, "book", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "magazine", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "printed_document", map[string]string{"name": "aa"}, 0)

		checkModelCount(t, client, "help_file", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "online_document", map[string]string{"name": "aa"}, 0)

		checkModelCount(t, client, "document", map[string]string{"name": "aa"}, 0)

		checkModelCount(t, client, "book", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "magazine", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "printed_document", map[string]string{"name": "cc"}, 8)

		checkModelCount(t, client, "help_file", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "cc"}, 4)
		checkModelCount(t, client, "online_document", map[string]string{"name": "cc"}, 8)

		checkModelCount(t, client, "document", map[string]string{"name": "cc"}, 16)
	})
}

func TestSTIByUpdateAllDocuments(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
		}
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
		}
		updateBy(t, client, "book", map[string]string{}, map[string]interface{}{"name": "bb"})

		checkModelCount(t, client, "book", map[string]string{"name": "bb"}, 8)
		checkModelCount(t, client, "book", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "book", map[string]string{"name": "cc"}, 0)

		checkModelCount(t, client, "magazine", map[string]string{"name": "bb"}, 0)
		checkModelCount(t, client, "magazine", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "magazine", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "help_file", map[string]string{"name": "bb"}, 0)
		checkModelCount(t, client, "help_file", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "help_file", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "bb"}, 0)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "cc"}, 4)
	})
}

func TestSTIByBaseClassUpdateAllDocuments(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
		}
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
		}

		updateBy(t, client, "printed_document", map[string]string{}, map[string]interface{}{"name": "bb"})

		checkModelCount(t, client, "book", map[string]string{"name": "bb"}, 8)
		checkModelCount(t, client, "book", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "book", map[string]string{"name": "cc"}, 0)

		checkModelCount(t, client, "magazine", map[string]string{"name": "bb"}, 8)
		checkModelCount(t, client, "magazine", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "magazine", map[string]string{"name": "cc"}, 0)

		checkModelCount(t, client, "help_file", map[string]string{"name": "bb"}, 0)
		checkModelCount(t, client, "help_file", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "help_file", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "bb"}, 0)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "cc"}, 4)
	})
}
func TestSTIByBaseClassUpdateAllDocuments2(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createDocument(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createPrintedDocument(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
		}
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createDocument(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createPrintedDocument(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
		}

		updateBy(t, client, "printed_document", map[string]string{}, map[string]interface{}{"name": "bb"})

		checkModelCount(t, client, "document", map[string]string{"name": "bb"}, 24)
		checkModelCount(t, client, "document", map[string]string{"name": "aa"}, 12)
		checkModelCount(t, client, "document", map[string]string{"name": "cc"}, 12)

		checkModelCount(t, client, "printed_document", map[string]string{"name": "bb"}, 24)
		checkModelCount(t, client, "printed_document", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "printed_document", map[string]string{"name": "cc"}, 0)

		checkModelCount(t, client, "book", map[string]string{"name": "bb"}, 8)
		checkModelCount(t, client, "book", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "book", map[string]string{"name": "cc"}, 0)

		checkModelCount(t, client, "magazine", map[string]string{"name": "bb"}, 8)
		checkModelCount(t, client, "magazine", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "magazine", map[string]string{"name": "cc"}, 0)

		checkModelCount(t, client, "help_file", map[string]string{"name": "bb"}, 0)
		checkModelCount(t, client, "help_file", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "help_file", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "bb"}, 0)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "cc"}, 4)
	})
}

func TestSTIByTopClassUpdateAllDocuments(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
		}
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
		}

		updateBy(t, client, "document", map[string]string{}, map[string]interface{}{"name": "bb"})

		checkModelCount(t, client, "book", map[string]string{"name": "bb"}, 8)
		checkModelCount(t, client, "book", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "book", map[string]string{"name": "cc"}, 0)

		checkModelCount(t, client, "magazine", map[string]string{"name": "bb"}, 8)
		checkModelCount(t, client, "magazine", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "magazine", map[string]string{"name": "cc"}, 0)

		checkModelCount(t, client, "help_file", map[string]string{"name": "bb"}, 8)
		checkModelCount(t, client, "help_file", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "help_file", map[string]string{"name": "cc"}, 0)

		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "bb"}, 8)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "cc"}, 0)
	})
}

func TestSTIByUpdateDocuments(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
		}
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
		}
		updateBy(t, client, "book", map[string]string{"name": "aa"}, map[string]interface{}{"name": "bb"})

		checkModelCount(t, client, "book", map[string]string{"name": "bb"}, 4)
		checkModelCount(t, client, "book", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "book", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "magazine", map[string]string{"name": "bb"}, 0)
		checkModelCount(t, client, "magazine", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "magazine", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "help_file", map[string]string{"name": "bb"}, 0)
		checkModelCount(t, client, "help_file", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "help_file", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "bb"}, 0)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "cc"}, 4)
	})
}

func TestSTIByBaseClassUpdateDocuments(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
		}
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
		}

		updateBy(t, client, "printed_document", map[string]string{"name": "aa"}, map[string]interface{}{"name": "bb"})

		checkModelCount(t, client, "book", map[string]string{"name": "bb"}, 4)
		checkModelCount(t, client, "book", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "book", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "magazine", map[string]string{"name": "bb"}, 4)
		checkModelCount(t, client, "magazine", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "magazine", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "help_file", map[string]string{"name": "bb"}, 0)
		checkModelCount(t, client, "help_file", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "help_file", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "bb"}, 0)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "aa"}, 4)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "cc"}, 4)
	})
}

func TestSTIByTopClassUpdateDocuments(t *testing.T) {
	SrvTest(t, "etc/mj_models.xml", func(client *Client, definitions *types.TableDefinitions) {
		clearAll(t, client)
		website1 := saveIt(t, client, "website", createWebsite())
		website2 := saveIt(t, client, "website", createWebsite())
		//website3 := saveIt(t, client, "website", createWebsite())

		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "cc"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "cc"}))
		}
		for i := 0; i < 4; i++ {
			saveIt(t, client, "document", change(createPrintedBook(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createMagazine(), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createHelpFile(website1), map[string]interface{}{"name": "aa"}))
			saveIt(t, client, "document", change(createOnlineTutorial(website2), map[string]interface{}{"name": "aa"}))
		}

		updateBy(t, client, "document", map[string]string{"name": "aa"}, map[string]interface{}{"name": "bb"})
		//     Document.update("name = ?", "name= ?", "bb", "aa");

		checkModelCount(t, client, "book", map[string]string{"name": "bb"}, 4)
		checkModelCount(t, client, "book", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "book", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "magazine", map[string]string{"name": "bb"}, 4)
		checkModelCount(t, client, "magazine", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "magazine", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "help_file", map[string]string{"name": "bb"}, 4)
		checkModelCount(t, client, "help_file", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "help_file", map[string]string{"name": "cc"}, 4)

		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "bb"}, 4)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "aa"}, 0)
		checkModelCount(t, client, "online_tutorial", map[string]string{"name": "cc"}, 4)
	})
}
