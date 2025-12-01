# Tutorial 1: Building Your First Service with Dependency Injection

This tutorial walks you through creating a complete service using the endor-sdk-go framework with dependency injection, from basic setup to production-ready implementation.

---

## What You'll Build

By the end of this tutorial, you'll have created a **Book Management Service** that demonstrates:

- ✅ Interface-driven dependency injection
- ✅ Unit testing with mocked dependencies  
- ✅ Integration testing with real database
- ✅ Proper error handling and logging
- ✅ Production-ready service configuration

**Estimated Time:** 30 minutes

---

## Prerequisites

- Go 1.21+ installed
- Basic understanding of Go interfaces
- MongoDB running locally (for integration tests)

---

## Step 1: Project Setup

### Initialize Your Project

```bash
mkdir book-service
cd book-service
go mod init github.com/yourname/book-service

# Add endor-sdk-go dependency
go get github.com/mattiabonardi/endor-sdk-go
```

### Create Basic Project Structure

```
book-service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── book/
│   │   ├── service.go
│   │   ├── service_test.go
│   │   ├── handler.go
│   │   └── model.go
│   └── config/
│       └── config.go
├── go.mod
└── README.md
```

---

## Step 2: Define Your Domain Model

Create `internal/book/model.go`:

```go
package book

import (
    "time"
)

// Book represents a book in our system
type Book struct {
    ID          string    `json:"id" bson:"_id"`
    Title       string    `json:"title" bson:"title"`
    Author      string    `json:"author" bson:"author"`
    ISBN        string    `json:"isbn" bson:"isbn"`
    PublishedAt time.Time `json:"published_at" bson:"published_at"`
    CreatedAt   time.Time `json:"created_at" bson:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

// CreateBookRequest represents the request to create a new book
type CreateBookRequest struct {
    Title       string    `json:"title" validate:"required,min=1,max=200"`
    Author      string    `json:"author" validate:"required,min=1,max=100"`
    ISBN        string    `json:"isbn" validate:"required,isbn"`
    PublishedAt time.Time `json:"published_at" validate:"required"`
}

// UpdateBookRequest represents the request to update a book
type UpdateBookRequest struct {
    Title       *string    `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
    Author      *string    `json:"author,omitempty" validate:"omitempty,min=1,max=100"`
    ISBN        *string    `json:"isbn,omitempty" validate:"omitempty,isbn"`
    PublishedAt *time.Time `json:"published_at,omitempty"`
}

// BookFilter represents filtering options for book queries
type BookFilter struct {
    Author    string `json:"author,omitempty"`
    Title     string `json:"title,omitempty"`
    ISBN      string `json:"isbn,omitempty"`
    Limit     int    `json:"limit,omitempty"`
    Offset    int    `json:"offset,omitempty"`
}
```

---

## Step 3: Create Service Interface

The key to dependency injection is starting with interfaces. Create `internal/book/interfaces.go`:

```go
package book

import (
    "context"
    "github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// BookServiceInterface defines the business logic interface for book management
type BookServiceInterface interface {
    CreateBook(ctx context.Context, request CreateBookRequest) (Book, error)
    GetBook(ctx context.Context, id string) (Book, error)
    UpdateBook(ctx context.Context, id string, request UpdateBookRequest) (Book, error)
    DeleteBook(ctx context.Context, id string) error
    ListBooks(ctx context.Context, filter BookFilter) ([]Book, error)
    SearchBooks(ctx context.Context, query string) ([]Book, error)
}

// BookRepositoryInterface defines the data access interface for books
type BookRepositoryInterface interface {
    Create(ctx context.Context, book Book) error
    FindByID(ctx context.Context, id string) (Book, error)
    Update(ctx context.Context, book Book) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter BookFilter) ([]Book, error)
    Search(ctx context.Context, query string) ([]Book, error)
    ExistsWithISBN(ctx context.Context, isbn string) (bool, error)
}

// BookValidatorInterface defines the validation interface for books
type BookValidatorInterface interface {
    ValidateCreateRequest(request CreateBookRequest) error
    ValidateUpdateRequest(request UpdateBookRequest) error
    ValidateISBN(isbn string) error
}
```

---

## Step 4: Implement the Service (Business Logic)

Create `internal/book/service.go`:

```go
package book

import (
    "context"
    "fmt"
    "time"
    "errors"
    
    "github.com/google/uuid"
    "github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
)

// BookService implements BookServiceInterface with dependency injection
type BookService struct {
    repository interfaces.RepositoryInterface // Generic repository interface
    validator  BookValidatorInterface
    logger     interfaces.LoggerInterface
}

// NewBookService creates a new BookService with injected dependencies
func NewBookService(
    repository interfaces.RepositoryInterface,
    validator BookValidatorInterface,
    logger interfaces.LoggerInterface,
) BookServiceInterface {
    return &BookService{
        repository: repository,
        validator:  validator,
        logger:     logger,
    }
}

func (s *BookService) CreateBook(ctx context.Context, request CreateBookRequest) (Book, error) {
    s.logger.Info("Creating new book", map[string]interface{}{
        "title":  request.Title,
        "author": request.Author,
        "isbn":   request.ISBN,
    })
    
    // Validate the request
    if err := s.validator.ValidateCreateRequest(request); err != nil {
        s.logger.Error("Book validation failed", err, map[string]interface{}{
            "title": request.Title,
            "isbn":  request.ISBN,
        })
        return Book{}, fmt.Errorf("validation failed: %w", err)
    }
    
    // Check for duplicate ISBN
    exists, err := s.checkISBNExists(ctx, request.ISBN)
    if err != nil {
        return Book{}, fmt.Errorf("failed to check ISBN uniqueness: %w", err)
    }
    if exists {
        s.logger.Error("Duplicate ISBN detected", nil, map[string]interface{}{
            "isbn": request.ISBN,
        })
        return Book{}, errors.New("book with this ISBN already exists")
    }
    
    // Create book entity
    now := time.Now()
    book := Book{
        ID:          uuid.New().String(),
        Title:       request.Title,
        Author:      request.Author,
        ISBN:        request.ISBN,
        PublishedAt: request.PublishedAt,
        CreatedAt:   now,
        UpdatedAt:   now,
    }
    
    // Persist to repository
    if err := s.repository.Create(ctx, book); err != nil {
        s.logger.Error("Failed to create book", err, map[string]interface{}{
            "book_id": book.ID,
            "isbn":    book.ISBN,
        })
        return Book{}, fmt.Errorf("failed to create book: %w", err)
    }
    
    s.logger.Info("Book created successfully", map[string]interface{}{
        "book_id": book.ID,
        "title":   book.Title,
        "isbn":    book.ISBN,
    })
    
    return book, nil
}

func (s *BookService) GetBook(ctx context.Context, id string) (Book, error) {
    s.logger.Debug("Retrieving book", map[string]interface{}{"book_id": id})
    
    var book Book
    if err := s.repository.FindByID(ctx, id, &book); err != nil {
        s.logger.Error("Failed to retrieve book", err, map[string]interface{}{
            "book_id": id,
        })
        return Book{}, fmt.Errorf("failed to retrieve book %s: %w", id, err)
    }
    
    return book, nil
}

func (s *BookService) UpdateBook(ctx context.Context, id string, request UpdateBookRequest) (Book, error) {
    s.logger.Info("Updating book", map[string]interface{}{"book_id": id})
    
    // Validate the request
    if err := s.validator.ValidateUpdateRequest(request); err != nil {
        s.logger.Error("Book update validation failed", err, map[string]interface{}{
            "book_id": id,
        })
        return Book{}, fmt.Errorf("validation failed: %w", err)
    }
    
    // Get existing book
    book, err := s.GetBook(ctx, id)
    if err != nil {
        return Book{}, err
    }
    
    // Apply updates
    if request.Title != nil {
        book.Title = *request.Title
    }
    if request.Author != nil {
        book.Author = *request.Author
    }
    if request.ISBN != nil {
        // Check for ISBN uniqueness if changed
        if book.ISBN != *request.ISBN {
            exists, err := s.checkISBNExists(ctx, *request.ISBN)
            if err != nil {
                return Book{}, fmt.Errorf("failed to check ISBN uniqueness: %w", err)
            }
            if exists {
                return Book{}, errors.New("book with this ISBN already exists")
            }
        }
        book.ISBN = *request.ISBN
    }
    if request.PublishedAt != nil {
        book.PublishedAt = *request.PublishedAt
    }
    
    book.UpdatedAt = time.Now()
    
    // Persist changes
    if err := s.repository.Update(ctx, book); err != nil {
        s.logger.Error("Failed to update book", err, map[string]interface{}{
            "book_id": id,
        })
        return Book{}, fmt.Errorf("failed to update book: %w", err)
    }
    
    s.logger.Info("Book updated successfully", map[string]interface{}{
        "book_id": id,
        "title":   book.Title,
    })
    
    return book, nil
}

func (s *BookService) DeleteBook(ctx context.Context, id string) error {
    s.logger.Info("Deleting book", map[string]interface{}{"book_id": id})
    
    // Verify book exists first
    _, err := s.GetBook(ctx, id)
    if err != nil {
        return err
    }
    
    if err := s.repository.Delete(ctx, id); err != nil {
        s.logger.Error("Failed to delete book", err, map[string]interface{}{
            "book_id": id,
        })
        return fmt.Errorf("failed to delete book: %w", err)
    }
    
    s.logger.Info("Book deleted successfully", map[string]interface{}{
        "book_id": id,
    })
    
    return nil
}

func (s *BookService) ListBooks(ctx context.Context, filter BookFilter) ([]Book, error) {
    s.logger.Debug("Listing books", map[string]interface{}{
        "filter": filter,
    })
    
    var books []Book
    if err := s.repository.FindAll(ctx, filter, &books); err != nil {
        s.logger.Error("Failed to list books", err, map[string]interface{}{
            "filter": filter,
        })
        return nil, fmt.Errorf("failed to list books: %w", err)
    }
    
    return books, nil
}

func (s *BookService) SearchBooks(ctx context.Context, query string) ([]Book, error) {
    s.logger.Debug("Searching books", map[string]interface{}{
        "query": query,
    })
    
    var books []Book
    if err := s.repository.Search(ctx, query, &books); err != nil {
        s.logger.Error("Failed to search books", err, map[string]interface{}{
            "query": query,
        })
        return nil, fmt.Errorf("failed to search books: %w", err)
    }
    
    return books, nil
}

// Helper method to check if ISBN already exists
func (s *BookService) checkISBNExists(ctx context.Context, isbn string) (bool, error) {
    var existingBook Book
    err := s.repository.FindByField(ctx, "isbn", isbn, &existingBook)
    if err != nil {
        if err.Error() == "not found" {
            return false, nil
        }
        return false, err
    }
    return true, nil
}
```

---

## Step 5: Create Test Setup

Create `internal/book/service_test.go`:

```go
package book

import (
    "context"
    "testing"
    "time"
    "errors"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/mattiabonardi/endor-sdk-go/sdk/testutils"
)

//go:build unit

func TestBookService_CreateBook_Success(t *testing.T) {
    // Arrange: Set up mocked dependencies
    mockRepo := testutils.NewMockRepository()
    mockValidator := NewMockBookValidator()
    mockLogger := testutils.NewMockLogger()
    
    request := CreateBookRequest{
        Title:       "The Go Programming Language",
        Author:      "Alan Donovan",
        ISBN:        "978-0134190440",
        PublishedAt: time.Date(2015, 10, 26, 0, 0, 0, 0, time.UTC),
    }
    
    // Configure mock expectations
    mockValidator.On("ValidateCreateRequest", request).Return(nil)
    mockRepo.On("FindByField", mock.Any, "isbn", request.ISBN, mock.Any).Return(errors.New("not found"))
    mockRepo.On("Create", mock.Any, mock.MatchedBy(func(book Book) bool {
        return book.Title == request.Title && book.ISBN == request.ISBN
    })).Return(nil)
    
    mockLogger.On("Info", "Creating new book", mock.Any).Return()
    mockLogger.On("Info", "Book created successfully", mock.Any).Return()
    mockLogger.On("Debug", mock.Any, mock.Any).Return()
    
    service := NewBookService(mockRepo, mockValidator, mockLogger)
    
    // Act: Execute the business logic
    book, err := service.CreateBook(context.Background(), request)
    
    // Assert: Verify behavior and mock interactions
    assert.NoError(t, err)
    assert.Equal(t, request.Title, book.Title)
    assert.Equal(t, request.Author, book.Author)
    assert.Equal(t, request.ISBN, book.ISBN)
    assert.NotEmpty(t, book.ID)
    assert.NotZero(t, book.CreatedAt)
    assert.Equal(t, book.CreatedAt, book.UpdatedAt)
    
    mockValidator.AssertExpectations(t)
    mockRepo.AssertExpectations(t)
    mockLogger.AssertExpectations(t)
}

func TestBookService_CreateBook_ValidationError(t *testing.T) {
    // Test validation error handling
    mockRepo := testutils.NewMockRepository()
    mockValidator := NewMockBookValidator()
    mockLogger := testutils.NewMockLogger()
    
    request := CreateBookRequest{
        Title: "", // Invalid: empty title
        ISBN:  "invalid-isbn",
    }
    
    validationError := errors.New("title is required")
    mockValidator.On("ValidateCreateRequest", request).Return(validationError)
    mockLogger.On("Info", "Creating new book", mock.Any).Return()
    mockLogger.On("Error", "Book validation failed", validationError, mock.Any).Return()
    
    service := NewBookService(mockRepo, mockValidator, mockLogger)
    
    book, err := service.CreateBook(context.Background(), request)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "validation failed")
    assert.Contains(t, err.Error(), "title is required")
    assert.Equal(t, Book{}, book)
    
    // Repository should not be called on validation failure
    mockRepo.AssertNotCalled(t, "Create")
    mockRepo.AssertNotCalled(t, "FindByField")
}

func TestBookService_CreateBook_DuplicateISBN(t *testing.T) {
    // Test duplicate ISBN handling
    mockRepo := testutils.NewMockRepository()
    mockValidator := NewMockBookValidator()
    mockLogger := testutils.NewMockLogger()
    
    request := CreateBookRequest{
        Title:       "Test Book",
        Author:      "Test Author",
        ISBN:        "978-0134190440",
        PublishedAt: time.Now(),
    }
    
    existingBook := Book{ID: "existing-id", ISBN: request.ISBN}
    
    mockValidator.On("ValidateCreateRequest", request).Return(nil)
    mockRepo.On("FindByField", mock.Any, "isbn", request.ISBN, mock.Any).Return(nil).Run(func(args mock.Arguments) {
        result := args.Get(3).(*Book)
        *result = existingBook
    })
    
    mockLogger.On("Info", "Creating new book", mock.Any).Return()
    mockLogger.On("Error", "Duplicate ISBN detected", mock.Any, mock.Any).Return()
    
    service := NewBookService(mockRepo, mockValidator, mockLogger)
    
    book, err := service.CreateBook(context.Background(), request)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "book with this ISBN already exists")
    assert.Equal(t, Book{}, book)
    
    // Create should not be called for duplicate ISBN
    mockRepo.AssertNotCalled(t, "Create")
}

func TestBookService_GetBook_Success(t *testing.T) {
    mockRepo := testutils.NewMockRepository()
    mockValidator := NewMockBookValidator()
    mockLogger := testutils.NewMockLogger()
    
    bookID := "test-book-id"
    expectedBook := Book{
        ID:     bookID,
        Title:  "Test Book",
        Author: "Test Author",
        ISBN:   "978-0134190440",
    }
    
    mockRepo.On("FindByID", mock.Any, bookID, mock.Any).Return(nil).Run(func(args mock.Arguments) {
        result := args.Get(2).(*Book)
        *result = expectedBook
    })
    mockLogger.On("Debug", "Retrieving book", mock.Any).Return()
    
    service := NewBookService(mockRepo, mockValidator, mockLogger)
    
    book, err := service.GetBook(context.Background(), bookID)
    
    assert.NoError(t, err)
    assert.Equal(t, expectedBook, book)
    mockRepo.AssertExpectations(t)
}

func TestBookService_GetBook_NotFound(t *testing.T) {
    mockRepo := testutils.NewMockRepository()
    mockValidator := NewMockBookValidator()
    mockLogger := testutils.NewMockLogger()
    
    bookID := "nonexistent-book"
    notFoundError := errors.New("book not found")
    
    mockRepo.On("FindByID", mock.Any, bookID, mock.Any).Return(notFoundError)
    mockLogger.On("Debug", "Retrieving book", mock.Any).Return()
    mockLogger.On("Error", "Failed to retrieve book", notFoundError, mock.Any).Return()
    
    service := NewBookService(mockRepo, mockValidator, mockLogger)
    
    book, err := service.GetBook(context.Background(), bookID)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to retrieve book")
    assert.Contains(t, err.Error(), "book not found")
    assert.Equal(t, Book{}, book)
}

// Mock implementations for testing
type MockBookValidator struct {
    mock.Mock
}

func NewMockBookValidator() *MockBookValidator {
    return &MockBookValidator{}
}

func (m *MockBookValidator) ValidateCreateRequest(request CreateBookRequest) error {
    args := m.Called(request)
    return args.Error(0)
}

func (m *MockBookValidator) ValidateUpdateRequest(request UpdateBookRequest) error {
    args := m.Called(request)
    return args.Error(0)
}

func (m *MockBookValidator) ValidateISBN(isbn string) error {
    args := m.Called(isbn)
    return args.Error(0)
}
```

---

## Step 6: Implement Production Validator

Create `internal/book/validator.go`:

```go
package book

import (
    "errors"
    "regexp"
    "strings"
    "time"
)

// ProductionBookValidator implements BookValidatorInterface for production use
type ProductionBookValidator struct{}

func NewProductionBookValidator() BookValidatorInterface {
    return &ProductionBookValidator{}
}

func (v *ProductionBookValidator) ValidateCreateRequest(request CreateBookRequest) error {
    if err := v.validateTitle(request.Title); err != nil {
        return err
    }
    
    if err := v.validateAuthor(request.Author); err != nil {
        return err
    }
    
    if err := v.ValidateISBN(request.ISBN); err != nil {
        return err
    }
    
    if err := v.validatePublishedAt(request.PublishedAt); err != nil {
        return err
    }
    
    return nil
}

func (v *ProductionBookValidator) ValidateUpdateRequest(request UpdateBookRequest) error {
    if request.Title != nil {
        if err := v.validateTitle(*request.Title); err != nil {
            return err
        }
    }
    
    if request.Author != nil {
        if err := v.validateAuthor(*request.Author); err != nil {
            return err
        }
    }
    
    if request.ISBN != nil {
        if err := v.ValidateISBN(*request.ISBN); err != nil {
            return err
        }
    }
    
    if request.PublishedAt != nil {
        if err := v.validatePublishedAt(*request.PublishedAt); err != nil {
            return err
        }
    }
    
    return nil
}

func (v *ProductionBookValidator) ValidateISBN(isbn string) error {
    // Remove hyphens and spaces
    cleanISBN := strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")
    
    // Check if it's a valid ISBN-10 or ISBN-13
    if len(cleanISBN) == 10 {
        return v.validateISBN10(cleanISBN)
    } else if len(cleanISBN) == 13 {
        return v.validateISBN13(cleanISBN)
    }
    
    return errors.New("ISBN must be 10 or 13 digits")
}

func (v *ProductionBookValidator) validateTitle(title string) error {
    title = strings.TrimSpace(title)
    if title == "" {
        return errors.New("title is required")
    }
    if len(title) < 1 {
        return errors.New("title must be at least 1 character")
    }
    if len(title) > 200 {
        return errors.New("title must be no more than 200 characters")
    }
    return nil
}

func (v *ProductionBookValidator) validateAuthor(author string) error {
    author = strings.TrimSpace(author)
    if author == "" {
        return errors.New("author is required")
    }
    if len(author) < 1 {
        return errors.New("author must be at least 1 character")
    }
    if len(author) > 100 {
        return errors.New("author must be no more than 100 characters")
    }
    return nil
}

func (v *ProductionBookValidator) validatePublishedAt(publishedAt time.Time) error {
    if publishedAt.IsZero() {
        return errors.New("published_at is required")
    }
    
    // Book can't be published in the future
    if publishedAt.After(time.Now()) {
        return errors.New("published_at cannot be in the future")
    }
    
    // Reasonable bounds: no books before 1400 AD
    if publishedAt.Year() < 1400 {
        return errors.New("published_at seems unreasonably old")
    }
    
    return nil
}

func (v *ProductionBookValidator) validateISBN10(isbn string) error {
    if !regexp.MustCompile(`^\d{9}[\dX]$`).MatchString(isbn) {
        return errors.New("invalid ISBN-10 format")
    }
    
    // ISBN-10 checksum validation
    sum := 0
    for i := 0; i < 9; i++ {
        digit := int(isbn[i] - '0')
        sum += digit * (10 - i)
    }
    
    checkDigit := isbn[9]
    if checkDigit == 'X' {
        sum += 10
    } else {
        sum += int(checkDigit - '0')
    }
    
    if sum%11 != 0 {
        return errors.New("invalid ISBN-10 checksum")
    }
    
    return nil
}

func (v *ProductionBookValidator) validateISBN13(isbn string) error {
    if !regexp.MustCompile(`^\d{13}$`).MatchString(isbn) {
        return errors.New("invalid ISBN-13 format")
    }
    
    // ISBN-13 checksum validation
    sum := 0
    for i := 0; i < 12; i++ {
        digit := int(isbn[i] - '0')
        if i%2 == 0 {
            sum += digit
        } else {
            sum += digit * 3
        }
    }
    
    checkDigit := int(isbn[12] - '0')
    expectedCheckDigit := (10 - (sum % 10)) % 10
    
    if checkDigit != expectedCheckDigit {
        return errors.New("invalid ISBN-13 checksum")
    }
    
    return nil
}
```

---

## Step 7: Set Up Dependency Injection

Create `cmd/server/main.go`:

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/mattiabonardi/endor-sdk-go/sdk"
    "github.com/mattiabonardi/endor-sdk-go/sdk/di"
    "github.com/mattiabonardi/endor-sdk-go/sdk/interfaces"
    
    "github.com/yourname/book-service/internal/book"
    "github.com/yourname/book-service/internal/config"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    // Set up dependency injection container
    container := di.NewContainer()
    if err := setupDependencies(container, cfg); err != nil {
        log.Fatalf("Failed to setup dependencies: %v", err)
    }
    
    // Validate all dependencies can be resolved
    if errors := container.Validate(); len(errors) > 0 {
        for _, err := range errors {
            log.Printf("Dependency validation error: %v", err)
        }
        log.Fatalf("Invalid dependency configuration")
    }
    
    // Resolve the book service
    bookService, err := di.Resolve[book.BookServiceInterface](container)
    if err != nil {
        log.Fatalf("Failed to resolve book service: %v", err)
    }
    
    // Set up HTTP server
    router := gin.New()
    router.Use(gin.Logger(), gin.Recovery())
    
    // Create book handler and register routes
    bookHandler := book.NewBookHandler(bookService)
    v1 := router.Group("/api/v1")
    {
        books := v1.Group("/books")
        {
            books.POST("", bookHandler.CreateBook)
            books.GET("", bookHandler.ListBooks)
            books.GET("/:id", bookHandler.GetBook)
            books.PUT("/:id", bookHandler.UpdateBook)
            books.DELETE("/:id", bookHandler.DeleteBook)
            books.GET("/search", bookHandler.SearchBooks)
        }
    }
    
    // Start server with graceful shutdown
    server := &http.Server{
        Addr:    cfg.ServerAddress,
        Handler: router,
    }
    
    go func() {
        log.Printf("Server starting on %s", cfg.ServerAddress)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed to start: %v", err)
        }
    }()
    
    // Wait for interrupt signal to gracefully shutdown the server
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit
    log.Println("Server is shutting down...")
    
    // Give outstanding requests a deadline for completion
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
    
    log.Println("Server exited")
}

func setupDependencies(container di.Container, cfg *config.Config) error {
    // Register configuration
    if err := di.Register[*config.Config](container, cfg, di.Singleton); err != nil {
        return err
    }
    
    // Register logger
    logger := &ProductionLogger{}
    if err := di.Register[interfaces.LoggerInterface](container, logger, di.Singleton); err != nil {
        return err
    }
    
    // Register MongoDB repository
    mongoRepo, err := NewMongoRepository(cfg.MongoURI, cfg.DatabaseName)
    if err != nil {
        return err
    }
    if err := di.Register[interfaces.RepositoryInterface](container, mongoRepo, di.Singleton); err != nil {
        return err
    }
    
    // Register book validator
    validator := book.NewProductionBookValidator()
    if err := di.Register[book.BookValidatorInterface](container, validator, di.Singleton); err != nil {
        return err
    }
    
    // Register book service factory
    if err := di.RegisterFactory[book.BookServiceInterface](container, func(c di.Container) (book.BookServiceInterface, error) {
        repo, err := di.Resolve[interfaces.RepositoryInterface](c)
        if err != nil {
            return nil, err
        }
        
        validator, err := di.Resolve[book.BookValidatorInterface](c)
        if err != nil {
            return nil, err
        }
        
        logger, err := di.Resolve[interfaces.LoggerInterface](c)
        if err != nil {
            return nil, err
        }
        
        return book.NewBookService(repo, validator, logger), nil
    }, di.Singleton); err != nil {
        return err
    }
    
    return nil
}

// ProductionLogger implements LoggerInterface for production use
type ProductionLogger struct{}

func (l *ProductionLogger) Debug(message string, fields map[string]interface{}) {
    log.Printf("[DEBUG] %s %+v", message, fields)
}

func (l *ProductionLogger) Info(message string, fields map[string]interface{}) {
    log.Printf("[INFO] %s %+v", message, fields)
}

func (l *ProductionLogger) Error(message string, err error, fields map[string]interface{}) {
    log.Printf("[ERROR] %s: %v %+v", message, err, fields)
}
```

---

## Step 8: Run and Test Your Service

### Run Unit Tests

```bash
# Run unit tests
go test -tags=unit ./internal/book/... -v

# Run with coverage
go test -tags=unit ./internal/book/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Start the Service

```bash
go run cmd/server/main.go
```

### Test the API

```bash
# Create a book
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "The Go Programming Language",
    "author": "Alan Donovan",
    "isbn": "978-0134190440",
    "published_at": "2015-10-26T00:00:00Z"
  }'

# Get all books
curl http://localhost:8080/api/v1/books

# Get specific book
curl http://localhost:8080/api/v1/books/{book-id}

# Search books
curl "http://localhost:8080/api/v1/books/search?q=Go"
```

---

## Step 9: Integration Testing

Create `internal/book/integration_test.go`:

```go
//go:build integration

package book

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/mattiabonardi/endor-sdk-go/sdk/di"
)

func TestBookService_Integration_CompleteWorkflow(t *testing.T) {
    // Set up integration test container
    container := di.NewContainer()
    setupIntegrationDependencies(container, t)
    
    // Resolve service with real implementations
    bookService, err := di.Resolve[BookServiceInterface](container)
    require.NoError(t, err)
    
    // Test complete workflow: Create → Read → Update → Delete
    ctx := context.Background()
    
    // 1. Create a book
    createRequest := CreateBookRequest{
        Title:       "Integration Test Book",
        Author:      "Integration Author",
        ISBN:        "978-1234567890",
        PublishedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
    }
    
    createdBook, err := bookService.CreateBook(ctx, createRequest)
    require.NoError(t, err)
    assert.Equal(t, createRequest.Title, createdBook.Title)
    assert.NotEmpty(t, createdBook.ID)
    
    // 2. Read the book back
    retrievedBook, err := bookService.GetBook(ctx, createdBook.ID)
    require.NoError(t, err)
    assert.Equal(t, createdBook.ID, retrievedBook.ID)
    assert.Equal(t, createdBook.Title, retrievedBook.Title)
    
    // 3. Update the book
    newTitle := "Updated Integration Test Book"
    updateRequest := UpdateBookRequest{
        Title: &newTitle,
    }
    
    updatedBook, err := bookService.UpdateBook(ctx, createdBook.ID, updateRequest)
    require.NoError(t, err)
    assert.Equal(t, newTitle, updatedBook.Title)
    assert.Equal(t, createdBook.Author, updatedBook.Author) // Unchanged
    
    // 4. List books (should include our book)
    books, err := bookService.ListBooks(ctx, BookFilter{})
    require.NoError(t, err)
    assert.Contains(t, books, updatedBook)
    
    // 5. Search for the book
    searchResults, err := bookService.SearchBooks(ctx, "Integration")
    require.NoError(t, err)
    assert.Contains(t, searchResults, updatedBook)
    
    // 6. Delete the book
    err = bookService.DeleteBook(ctx, createdBook.ID)
    require.NoError(t, err)
    
    // 7. Verify book is deleted
    _, err = bookService.GetBook(ctx, createdBook.ID)
    assert.Error(t, err) // Should not find deleted book
}

func TestBookService_Integration_DuplicateISBN(t *testing.T) {
    container := di.NewContainer()
    setupIntegrationDependencies(container, t)
    
    bookService, err := di.Resolve[BookServiceInterface](container)
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Create first book
    createRequest := CreateBookRequest{
        Title:       "First Book",
        Author:      "First Author",
        ISBN:        "978-1111111111",
        PublishedAt: time.Now(),
    }
    
    _, err = bookService.CreateBook(ctx, createRequest)
    require.NoError(t, err)
    
    // Try to create second book with same ISBN
    duplicateRequest := CreateBookRequest{
        Title:       "Duplicate Book",
        Author:      "Duplicate Author",
        ISBN:        "978-1111111111", // Same ISBN
        PublishedAt: time.Now(),
    }
    
    _, err = bookService.CreateBook(ctx, duplicateRequest)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "book with this ISBN already exists")
}
```

---

## What You've Accomplished 

✅ **Created a production-ready service** with clean architecture  
✅ **Implemented dependency injection** for testability and flexibility  
✅ **Wrote comprehensive unit tests** with mocked dependencies  
✅ **Created integration tests** with real database operations  
✅ **Applied proper error handling** and logging throughout  
✅ **Used interfaces** for all external dependencies  

### Key Benefits You've Gained

1. **Testable**: Unit tests run fast without external dependencies
2. **Maintainable**: Clear separation of concerns and dependency injection
3. **Flexible**: Easy to swap implementations (e.g., different databases)
4. **Robust**: Comprehensive validation and error handling
5. **Production-ready**: Proper logging, configuration, and graceful shutdown

---

## Next Steps

Ready to level up? Here's what to explore next:

1. **[Tutorial 2: Advanced Composition](tutorial-2-composition.md)** - Service composition and embedding
2. **[Tutorial 3: Testing Strategies](tutorial-3-testing.md)** - Advanced testing patterns  
3. **[Tutorial 4: Performance Optimization](tutorial-4-performance.md)** - Optimization techniques

### Additional Features to Add

- **Pagination**: Add limit/offset to ListBooks
- **Caching**: Add Redis caching layer
- **Authentication**: Add JWT authentication middleware
- **Rate Limiting**: Add request rate limiting
- **Metrics**: Add Prometheus metrics
- **Docker**: Containerize your service

Congratulations on building your first endor-sdk-go service! 🎉