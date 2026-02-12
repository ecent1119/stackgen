package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Generate test containers and test function scaffolding",
	Long: `Generate test containers and test function scaffolding for your project.

stackgen test provides a TUI for creating:
- Test containers (Go, Node, Python, Java, Rust, C#)
- Integration test scaffolding
- Test environment configuration

Examples:
  stackgen test              # Launch TUI
  stackgen test --runtime go # Generate Go test container`,
	RunE: runTest,
}

var testRuntime string

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().StringVarP(&testRuntime, "runtime", "r", "", "runtime for test container (go, node, python, java, rust, csharp)")
}

// TUI Model
type testModel struct {
	step       int
	runtime    string
	testType   string
	outputDir  string
	list       list.Model
	textInput  textinput.Model
	err        error
	done       bool
	generated  *testOutput
}

type testOutput struct {
	Dockerfile    string
	ComposeAdd    string
	TestFile      string
	TestFileName  string
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("10"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

func initialTestModel() testModel {
	// Runtime selection list
	items := []list.Item{
		item{title: "go", desc: "Go test container with go test"},
		item{title: "node", desc: "Node.js test container with Jest/Vitest"},
		item{title: "python", desc: "Python test container with pytest"},
		item{title: "java", desc: "Java test container with JUnit"},
		item{title: "rust", desc: "Rust test container with cargo test"},
		item{title: "csharp", desc: "C# test container with xUnit"},
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 50, 14)
	l.Title = "Select test runtime"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle

	ti := textinput.New()
	ti.Placeholder = "."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return testModel{
		step:      0,
		list:      l,
		textInput: ti,
	}
}

func (m testModel) Init() tea.Cmd {
	return nil
}

func (m testModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			switch m.step {
			case 0: // Runtime selected
				if i, ok := m.list.SelectedItem().(item); ok {
					m.runtime = i.title
					m.step = 1
					// Update list for test type
					items := []list.Item{
						item{title: "unit", desc: "Unit tests for isolated functions"},
						item{title: "integration", desc: "Integration tests with services"},
						item{title: "e2e", desc: "End-to-end tests"},
					}
					m.list.SetItems(items)
					m.list.Title = "Select test type"
				}
			case 1: // Test type selected
				if i, ok := m.list.SelectedItem().(item); ok {
					m.testType = i.title
					m.step = 2
				}
			case 2: // Output dir entered
				m.outputDir = m.textInput.Value()
				if m.outputDir == "" {
					m.outputDir = "."
				}
				m.generated = generateTestOutput(m.runtime, m.testType, m.outputDir)
				m.done = true
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	}

	var cmd tea.Cmd
	if m.step < 2 {
		m.list, cmd = m.list.Update(msg)
	} else {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

func (m testModel) View() string {
	if m.done {
		return ""
	}

	var s strings.Builder
	s.WriteString(titleStyle.Render("stackgen test — Test Container Generator"))
	s.WriteString("\n\n")

	switch m.step {
	case 0, 1:
		s.WriteString(m.list.View())
	case 2:
		s.WriteString(fmt.Sprintf("Runtime: %s\n", selectedStyle.Render(m.runtime)))
		s.WriteString(fmt.Sprintf("Test type: %s\n\n", selectedStyle.Render(m.testType)))
		s.WriteString("Output directory:\n")
		s.WriteString(m.textInput.View())
	}

	s.WriteString(helpStyle.Render("\n↑/↓: navigate • enter: select • q: quit"))
	return s.String()
}

func runTest(cmd *cobra.Command, args []string) error {
	// Non-interactive mode
	if testRuntime != "" {
		output := generateTestOutput(testRuntime, "integration", ".")
		if output == nil {
			return fmt.Errorf("unsupported runtime: %s", testRuntime)
		}
		return writeTestOutput(output, ".")
	}

	// TUI mode
	p := tea.NewProgram(initialTestModel())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	m, ok := finalModel.(testModel)
	if !ok || m.generated == nil {
		return nil
	}

	return writeTestOutput(m.generated, m.outputDir)
}

func generateTestOutput(runtime, testType, outputDir string) *testOutput {
	output := &testOutput{}

	switch runtime {
	case "go":
		output.Dockerfile = goTestDockerfile()
		output.ComposeAdd = goTestCompose(testType)
		output.TestFile = goTestFile(testType)
		output.TestFileName = "main_test.go"
	case "node":
		output.Dockerfile = nodeTestDockerfile()
		output.ComposeAdd = nodeTestCompose(testType)
		output.TestFile = nodeTestFile(testType)
		output.TestFileName = "test/app.test.js"
	case "python":
		output.Dockerfile = pythonTestDockerfile()
		output.ComposeAdd = pythonTestCompose(testType)
		output.TestFile = pythonTestFile(testType)
		output.TestFileName = "tests/test_app.py"
	case "java":
		output.Dockerfile = javaTestDockerfile()
		output.ComposeAdd = javaTestCompose(testType)
		output.TestFile = javaTestFile(testType)
		output.TestFileName = "src/test/java/AppTest.java"
	case "rust":
		output.Dockerfile = rustTestDockerfile()
		output.ComposeAdd = rustTestCompose(testType)
		output.TestFile = rustTestFile(testType)
		output.TestFileName = "tests/integration_test.rs"
	case "csharp":
		output.Dockerfile = csharpTestDockerfile()
		output.ComposeAdd = csharpTestCompose(testType)
		output.TestFile = csharpTestFile(testType)
		output.TestFileName = "Tests/AppTests.cs"
	default:
		return nil
	}

	return output
}

func writeTestOutput(output *testOutput, outputDir string) error {
	absDir, _ := filepath.Abs(outputDir)

	// Create test directory
	testDir := filepath.Join(absDir, "test-container")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		return err
	}

	// Write Dockerfile
	dockerPath := filepath.Join(testDir, "Dockerfile.test")
	if err := os.WriteFile(dockerPath, []byte(output.Dockerfile), 0644); err != nil {
		return err
	}

	// Write compose addition
	composePath := filepath.Join(testDir, "docker-compose.test.yml")
	if err := os.WriteFile(composePath, []byte(output.ComposeAdd), 0644); err != nil {
		return err
	}

	// Write test file template
	testFilePath := filepath.Join(testDir, filepath.Base(output.TestFileName))
	if err := os.WriteFile(testFilePath, []byte(output.TestFile), 0644); err != nil {
		return err
	}

	color.Green("\n✅ Test scaffolding generated!\n\n")
	fmt.Println("Generated files:")
	fmt.Printf("  • %s\n", color.CyanString("test-container/Dockerfile.test"))
	fmt.Printf("  • %s\n", color.CyanString("test-container/docker-compose.test.yml"))
	fmt.Printf("  • %s\n", color.CyanString("test-container/"+filepath.Base(output.TestFileName)))

	fmt.Println("\nUsage:")
	color.Yellow("  # Run tests in container")
	color.Yellow("  docker compose -f docker-compose.yml -f test-container/docker-compose.test.yml run --rm test")
	fmt.Println()

	return nil
}

// Go test templates
func goTestDockerfile() string {
	return `# Go Test Container - Generated by stackgen
FROM golang:1.22-alpine

WORKDIR /app

# Install test dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Run tests
CMD ["go", "test", "-v", "-race", "-coverprofile=coverage.out", "./..."]
`
}

func goTestCompose(testType string) string {
	compose := `# Go Test Service - Generated by stackgen
# Add to your docker-compose.yml or use with -f flag

services:
  test:
    build:
      context: .
      dockerfile: test-container/Dockerfile.test
    volumes:
      - .:/app
    environment:
      - CGO_ENABLED=1
`
	if testType == "integration" {
		compose += `    depends_on:
      - postgres
      - redis
    env_file:
      - .env
`
	}
	return compose
}

func goTestFile(testType string) string {
	if testType == "integration" {
		return `package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	// Setup
	code := m.Run()
	// Teardown
	os.Exit(code)
}

func TestDatabaseConnection(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping: %v", err)
	}
}

func TestExample(t *testing.T) {
	// Your test here
	t.Log("Test passed")
}
`
	}
	return `package main

import "testing"

func TestExample(t *testing.T) {
	// Your unit test here
	got := 1 + 1
	want := 2
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestAnotherExample(t *testing.T) {
	// Another test
	t.Log("Test passed")
}
`
}

// Node test templates
func nodeTestDockerfile() string {
	return `# Node.js Test Container - Generated by stackgen
FROM node:20-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci

# Copy source
COPY . .

# Run tests
CMD ["npm", "test"]
`
}

func nodeTestCompose(testType string) string {
	compose := `# Node.js Test Service - Generated by stackgen
services:
  test:
    build:
      context: .
      dockerfile: test-container/Dockerfile.test
    volumes:
      - .:/app
      - /app/node_modules
    environment:
      - NODE_ENV=test
`
	if testType == "integration" {
		compose += `    depends_on:
      - postgres
      - redis
    env_file:
      - .env
`
	}
	return compose
}

func nodeTestFile(testType string) string {
	if testType == "integration" {
		return `// Integration tests - Generated by stackgen
const { Pool } = require('pg');

describe('Database Integration', () => {
  let pool;

  beforeAll(async () => {
    pool = new Pool({
      connectionString: process.env.DATABASE_URL
    });
  });

  afterAll(async () => {
    await pool.end();
  });

  test('connects to database', async () => {
    const result = await pool.query('SELECT 1 as value');
    expect(result.rows[0].value).toBe(1);
  });

  test('example integration test', async () => {
    // Your integration test here
    expect(true).toBe(true);
  });
});
`
	}
	return `// Unit tests - Generated by stackgen

describe('Example Tests', () => {
  test('adds 1 + 1 to equal 2', () => {
    expect(1 + 1).toBe(2);
  });

  test('example async test', async () => {
    const result = await Promise.resolve('success');
    expect(result).toBe('success');
  });
});
`
}

// Python test templates
func pythonTestDockerfile() string {
	return `# Python Test Container - Generated by stackgen
FROM python:3.12-slim

WORKDIR /app

# Install test dependencies
COPY requirements*.txt ./
RUN pip install --no-cache-dir -r requirements.txt || true
RUN pip install pytest pytest-cov pytest-asyncio

# Copy source
COPY . .

# Run tests
CMD ["pytest", "-v", "--cov=.", "--cov-report=term-missing"]
`
}

func pythonTestCompose(testType string) string {
	compose := `# Python Test Service - Generated by stackgen
services:
  test:
    build:
      context: .
      dockerfile: test-container/Dockerfile.test
    volumes:
      - .:/app
    environment:
      - PYTHONPATH=/app
`
	if testType == "integration" {
		compose += `    depends_on:
      - postgres
      - redis
    env_file:
      - .env
`
	}
	return compose
}

func pythonTestFile(testType string) string {
	if testType == "integration" {
		return `"""Integration tests - Generated by stackgen"""
import os
import pytest
import psycopg2


@pytest.fixture
def db_connection():
    """Database connection fixture."""
    conn = psycopg2.connect(os.environ.get("DATABASE_URL"))
    yield conn
    conn.close()


def test_database_connection(db_connection):
    """Test database connectivity."""
    cursor = db_connection.cursor()
    cursor.execute("SELECT 1")
    result = cursor.fetchone()
    assert result[0] == 1


def test_example_integration():
    """Example integration test."""
    # Your integration test here
    assert True
`
	}
	return `"""Unit tests - Generated by stackgen"""
import pytest


def test_example():
    """Example unit test."""
    assert 1 + 1 == 2


def test_another_example():
    """Another example test."""
    result = "hello".upper()
    assert result == "HELLO"


@pytest.mark.asyncio
async def test_async_example():
    """Example async test."""
    import asyncio
    await asyncio.sleep(0.1)
    assert True
`
}

// Java test templates
func javaTestDockerfile() string {
	return `# Java Test Container - Generated by stackgen
FROM eclipse-temurin:21-jdk-alpine

WORKDIR /app

# Copy build files
COPY pom.xml* mvnw* ./
COPY .mvn* .mvn/
COPY build.gradle* gradlew* ./
COPY gradle* gradle/

# Download dependencies
RUN if [ -f mvnw ]; then ./mvnw dependency:go-offline; \
    elif [ -f gradlew ]; then ./gradlew dependencies; fi || true

# Copy source
COPY . .

# Run tests
CMD ["sh", "-c", "if [ -f mvnw ]; then ./mvnw test; elif [ -f gradlew ]; then ./gradlew test; fi"]
`
}

func javaTestCompose(testType string) string {
	compose := `# Java Test Service - Generated by stackgen
services:
  test:
    build:
      context: .
      dockerfile: test-container/Dockerfile.test
    volumes:
      - .:/app
      - maven-cache:/root/.m2
`
	if testType == "integration" {
		compose += `    depends_on:
      - postgres
      - redis
    env_file:
      - .env
`
	}
	compose += `
volumes:
  maven-cache:
`
	return compose
}

func javaTestFile(testType string) string {
	if testType == "integration" {
		return `// Integration tests - Generated by stackgen
package com.example;

import org.junit.jupiter.api.*;
import java.sql.*;

import static org.junit.jupiter.api.Assertions.*;

class IntegrationTest {

    private static Connection connection;

    @BeforeAll
    static void setUp() throws SQLException {
        String dbUrl = System.getenv("DATABASE_URL");
        if (dbUrl != null) {
            connection = DriverManager.getConnection(dbUrl);
        }
    }

    @AfterAll
    static void tearDown() throws SQLException {
        if (connection != null) {
            connection.close();
        }
    }

    @Test
    void testDatabaseConnection() throws SQLException {
        Assumptions.assumeTrue(connection != null, "Database not configured");
        try (Statement stmt = connection.createStatement()) {
            ResultSet rs = stmt.executeQuery("SELECT 1");
            assertTrue(rs.next());
            assertEquals(1, rs.getInt(1));
        }
    }

    @Test
    void testExample() {
        // Your integration test here
        assertTrue(true);
    }
}
`
	}
	return `// Unit tests - Generated by stackgen
package com.example;

import org.junit.jupiter.api.*;
import static org.junit.jupiter.api.Assertions.*;

class AppTest {

    @Test
    void testAddition() {
        assertEquals(2, 1 + 1);
    }

    @Test
    void testString() {
        String result = "hello".toUpperCase();
        assertEquals("HELLO", result);
    }

    @Test
    void testExample() {
        // Your unit test here
        assertTrue(true);
    }
}
`
}

// Rust test templates
func rustTestDockerfile() string {
	return `# Rust Test Container - Generated by stackgen
FROM rust:1.75-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache musl-dev

# Copy manifests
COPY Cargo.toml Cargo.lock* ./

# Create dummy src for dependency caching
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release || true
RUN rm -rf src

# Copy source
COPY . .

# Run tests
CMD ["cargo", "test", "--", "--nocapture"]
`
}

func rustTestCompose(testType string) string {
	compose := `# Rust Test Service - Generated by stackgen
services:
  test:
    build:
      context: .
      dockerfile: test-container/Dockerfile.test
    volumes:
      - .:/app
      - cargo-cache:/usr/local/cargo/registry
`
	if testType == "integration" {
		compose += `    depends_on:
      - postgres
      - redis
    env_file:
      - .env
`
	}
	compose += `
volumes:
  cargo-cache:
`
	return compose
}

func rustTestFile(testType string) string {
	if testType == "integration" {
		return `// Integration tests - Generated by stackgen
use std::env;

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_database_url_configured() {
        // Check if DATABASE_URL is set
        if env::var("DATABASE_URL").is_err() {
            println!("DATABASE_URL not set, skipping");
            return;
        }
        // Your integration test here
        assert!(true);
    }

    #[test]
    fn test_example_integration() {
        // Your integration test here
        assert_eq!(1 + 1, 2);
    }
}
`
	}
	return `// Unit tests - Generated by stackgen

#[cfg(test)]
mod tests {
    #[test]
    fn test_addition() {
        assert_eq!(1 + 1, 2);
    }

    #[test]
    fn test_string() {
        let result = "hello".to_uppercase();
        assert_eq!(result, "HELLO");
    }

    #[test]
    fn test_example() {
        // Your unit test here
        assert!(true);
    }
}
`
}

// C# test templates
func csharpTestDockerfile() string {
	return `# C# Test Container - Generated by stackgen
FROM mcr.microsoft.com/dotnet/sdk:8.0-alpine

WORKDIR /app

# Copy project files
COPY *.csproj *.sln ./
RUN dotnet restore || true

# Copy source
COPY . .

# Run tests
CMD ["dotnet", "test", "--verbosity", "normal"]
`
}

func csharpTestCompose(testType string) string {
	compose := `# C# Test Service - Generated by stackgen
services:
  test:
    build:
      context: .
      dockerfile: test-container/Dockerfile.test
    volumes:
      - .:/app
`
	if testType == "integration" {
		compose += `    depends_on:
      - mssql
    env_file:
      - .env
`
	}
	return compose
}

func csharpTestFile(testType string) string {
	if testType == "integration" {
		return `// Integration tests - Generated by stackgen
using Xunit;
using System;
using System.Data.SqlClient;

namespace Tests;

public class IntegrationTests
{
    [Fact]
    public void TestDatabaseConnection()
    {
        var connectionString = Environment.GetEnvironmentVariable("MSSQL_URL");
        if (string.IsNullOrEmpty(connectionString))
        {
            // Skip if not configured
            return;
        }

        using var connection = new SqlConnection(connectionString);
        connection.Open();

        using var command = new SqlCommand("SELECT 1", connection);
        var result = command.ExecuteScalar();

        Assert.Equal(1, result);
    }

    [Fact]
    public void TestExample()
    {
        // Your integration test here
        Assert.True(true);
    }
}
`
	}
	return `// Unit tests - Generated by stackgen
using Xunit;

namespace Tests;

public class AppTests
{
    [Fact]
    public void TestAddition()
    {
        Assert.Equal(2, 1 + 1);
    }

    [Fact]
    public void TestString()
    {
        var result = "hello".ToUpper();
        Assert.Equal("HELLO", result);
    }

    [Fact]
    public void TestExample()
    {
        // Your unit test here
        Assert.True(true);
    }
}
`
}
