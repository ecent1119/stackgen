package models

import "testing"

func TestAvailableDatastores(t *testing.T) {
	datastores := AvailableDatastores()
	
	expected := 6
	if len(datastores) != expected {
		t.Errorf("Expected %d datastores, got %d", expected, len(datastores))
	}

	// Check all types are present
	types := map[DatastoreType]bool{
		DatastorePostgres:   false,
		DatastoreMySQL:      false,
		DatastoreMSSQL:      false,
		DatastoreNeo4j:      false,
		DatastoreRedis:      false,
		DatastoreRedisStack: false,
	}

	for _, ds := range datastores {
		types[ds] = true
	}

	for dsType, found := range types {
		if !found {
			t.Errorf("Datastore type %s not found", dsType)
		}
	}
}

func TestAvailableRuntimes(t *testing.T) {
	runtimes := AvailableRuntimes()
	
	expected := 6
	if len(runtimes) != expected {
		t.Errorf("Expected %d runtimes, got %d", expected, len(runtimes))
	}

	// Check all types are present
	types := map[RuntimeType]bool{
		RuntimeGo:     false,
		RuntimeNode:   false,
		RuntimePython: false,
		RuntimeJava:   false,
		RuntimeRust:   false,
		RuntimeCSharp: false,
	}

	for _, rt := range runtimes {
		types[rt] = true
	}

	for rtType, found := range types {
		if !found {
			t.Errorf("Runtime type %s not found", rtType)
		}
	}
}

func TestGetDatastoreInfo(t *testing.T) {
	info := GetDatastoreInfo(DatastorePostgres)
	
	if info.DisplayName != "PostgreSQL" {
		t.Errorf("Expected PostgreSQL, got %s", info.DisplayName)
	}

	if info.DefaultPort != 5432 {
		t.Errorf("Expected port 5432, got %d", info.DefaultPort)
	}
}

func TestGetRuntimeInfo(t *testing.T) {
	info := GetRuntimeInfo(RuntimeNode)
	
	if info.DisplayName != "Node.js" {
		t.Errorf("Expected Node.js, got %s", info.DisplayName)
	}

	if info.DefaultPort != 3000 {
		t.Errorf("Expected port 3000, got %d", info.DefaultPort)
	}

	if len(info.Frameworks) == 0 {
		t.Error("Expected frameworks to be defined")
	}
}

func TestMSSQLInfo(t *testing.T) {
	info := GetDatastoreInfo(DatastoreMSSQL)
	
	// Verify Developer Edition is specified
	if info.Edition != "Developer Edition - for development use only" {
		t.Error("MSSQL should specify Developer Edition")
	}
}

func TestNeo4jInfo(t *testing.T) {
	info := GetDatastoreInfo(DatastoreNeo4j)
	
	// Verify Community Edition is specified
	if info.Edition != "Community Edition" {
		t.Error("Neo4j should specify Community Edition")
	}
}
