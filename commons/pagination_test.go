package commons

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PaginationTestSuite struct {
	suite.Suite
}

func TestPaginationTestSuite(t *testing.T) {
	suite.Run(t, new(PaginationTestSuite))
}

// Tests for NewPagination constructor
func (suite *PaginationTestSuite) TestNewPagination_Success() {
	// Arrange
	page := 2
	pageSize := 20
	orderBy := "created_at"
	sort := OrderAsc
	query := "test query"

	// Act
	p := NewPagination(page, pageSize, orderBy, sort, query)

	// Assert
	assert.NotNil(suite.T(), p)
	assert.Equal(suite.T(), page, p.GetPage())
	assert.Equal(suite.T(), pageSize, p.GetPageSize())
	assert.Equal(suite.T(), orderBy, p.GetOrderBy())
	assert.Equal(suite.T(), sort, p.GetSort())
	assert.Equal(suite.T(), query, p.GetQuery())
}

// Tests for GetPage
func (suite *PaginationTestSuite) TestGetPage_ValidPage() {
	// Arrange
	p := NewPagination(5, 10, "", OrderAsc, "")

	// Act
	page := p.GetPage()

	// Assert
	assert.Equal(suite.T(), 5, page)
}

func (suite *PaginationTestSuite) TestGetPage_ZeroReturnsDefault() {
	// Arrange
	p := NewPagination(0, 10, "", OrderAsc, "")

	// Act
	page := p.GetPage()

	// Assert
	assert.Equal(suite.T(), 1, page)
}

func (suite *PaginationTestSuite) TestGetPage_NegativeReturnsDefault() {
	// Arrange
	p := NewPagination(-5, 10, "", OrderAsc, "")

	// Act
	page := p.GetPage()

	// Assert
	assert.Equal(suite.T(), 1, page)
}

// Tests for GetPageSize
func (suite *PaginationTestSuite) TestGetPageSize_ValidPageSize() {
	// Arrange
	p := NewPagination(1, 25, "", OrderAsc, "")

	// Act
	pageSize := p.GetPageSize()

	// Assert
	assert.Equal(suite.T(), 25, pageSize)
}

func (suite *PaginationTestSuite) TestGetPageSize_ZeroReturnsDefault() {
	// Arrange
	p := NewPagination(1, 0, "", OrderAsc, "")

	// Act
	pageSize := p.GetPageSize()

	// Assert
	assert.Equal(suite.T(), 10, pageSize)
}

func (suite *PaginationTestSuite) TestGetPageSize_NegativeReturnsDefault() {
	// Arrange
	p := NewPagination(1, -10, "", OrderAsc, "")

	// Act
	pageSize := p.GetPageSize()

	// Assert
	assert.Equal(suite.T(), 10, pageSize)
}

// Tests for GetOrderBy
func (suite *PaginationTestSuite) TestGetOrderBy_ValidOrderBy() {
	// Arrange
	p := NewPagination(1, 10, "name", OrderAsc, "")

	// Act
	orderBy := p.GetOrderBy()

	// Assert
	assert.Equal(suite.T(), "name", orderBy)
}

func (suite *PaginationTestSuite) TestGetOrderBy_EmptyReturnsDefault() {
	// Arrange
	p := NewPagination(1, 10, "", OrderAsc, "")

	// Act
	orderBy := p.GetOrderBy()

	// Assert
	assert.Equal(suite.T(), "updated_at", orderBy)
}

// Tests for GetSort
func (suite *PaginationTestSuite) TestGetSort_ValidSortAsc() {
	// Arrange
	p := NewPagination(1, 10, "", OrderAsc, "")

	// Act
	sort := p.GetSort()

	// Assert
	assert.Equal(suite.T(), OrderAsc, sort)
}

func (suite *PaginationTestSuite) TestGetSort_ValidSortDesc() {
	// Arrange
	p := NewPagination(1, 10, "", OrderDesc, "")

	// Act
	sort := p.GetSort()

	// Assert
	assert.Equal(suite.T(), OrderDesc, sort)
}

func (suite *PaginationTestSuite) TestGetSort_EmptyReturnsDefault() {
	// Arrange
	p := NewPagination(1, 10, "", "", "")

	// Act
	sort := p.GetSort()

	// Assert
	assert.Equal(suite.T(), OrderDesc, sort)
}

// Tests for GetQuery
func (suite *PaginationTestSuite) TestGetQuery_ValidQuery() {
	// Arrange
	query := "search term"
	p := NewPagination(1, 10, "", OrderAsc, query)

	// Act
	result := p.GetQuery()

	// Assert
	assert.Equal(suite.T(), query, result)
}

func (suite *PaginationTestSuite) TestGetQuery_EmptyQuery() {
	// Arrange
	p := NewPagination(1, 10, "", OrderAsc, "")

	// Act
	result := p.GetQuery()

	// Assert
	assert.Equal(suite.T(), "", result)
}

// Tests for Offset
func (suite *PaginationTestSuite) TestOffset_FirstPage() {
	// Arrange
	p := NewPagination(1, 10, "", OrderAsc, "")

	// Act
	offset := p.Offset()

	// Assert
	assert.Equal(suite.T(), 0, offset)
}

func (suite *PaginationTestSuite) TestOffset_SecondPage() {
	// Arrange
	p := NewPagination(2, 10, "", OrderAsc, "")

	// Act
	offset := p.Offset()

	// Assert
	assert.Equal(suite.T(), 10, offset)
}

func (suite *PaginationTestSuite) TestOffset_ThirdPageWithCustomPageSize() {
	// Arrange
	p := NewPagination(3, 25, "", OrderAsc, "")

	// Act
	offset := p.Offset()

	// Assert
	assert.Equal(suite.T(), 50, offset)
}

func (suite *PaginationTestSuite) TestOffset_WithDefaultPage() {
	// Arrange - page 0 should default to 1
	p := NewPagination(0, 10, "", OrderAsc, "")

	// Act
	offset := p.Offset()

	// Assert
	assert.Equal(suite.T(), 0, offset)
}

// Tests for Limit
func (suite *PaginationTestSuite) TestLimit_ValidPageSize() {
	// Arrange
	p := NewPagination(1, 20, "", OrderAsc, "")

	// Act
	limit := p.Limit()

	// Assert
	assert.Equal(suite.T(), 20, limit)
}

func (suite *PaginationTestSuite) TestLimit_DefaultPageSize() {
	// Arrange
	p := NewPagination(1, 0, "", OrderAsc, "")

	// Act
	limit := p.Limit()

	// Assert
	assert.Equal(suite.T(), 10, limit)
}

// Tests for Order
func (suite *PaginationTestSuite) TestOrder_WithAscending() {
	// Arrange
	p := NewPagination(1, 10, "name", OrderAsc, "")

	// Act
	order := p.Order()

	// Assert
	assert.Equal(suite.T(), "name asc", order)
}

func (suite *PaginationTestSuite) TestOrder_WithDescending() {
	// Arrange
	p := NewPagination(1, 10, "created_at", OrderDesc, "")

	// Act
	order := p.Order()

	// Assert
	assert.Equal(suite.T(), "created_at desc", order)
}

func (suite *PaginationTestSuite) TestOrder_WithDefaults() {
	// Arrange
	p := NewPagination(1, 10, "", "", "")

	// Act
	order := p.Order()

	// Assert
	assert.Equal(suite.T(), "updated_at desc", order)
}

// Tests for GetTotalPages
func (suite *PaginationTestSuite) TestGetTotalPages_ExactDivision() {
	// Arrange
	p := NewPagination(1, 10, "", OrderAsc, "")

	// Act
	totalPages := p.GetTotalPages(100)

	// Assert
	assert.Equal(suite.T(), 10, totalPages)
}

func (suite *PaginationTestSuite) TestGetTotalPages_WithRemainder() {
	// Arrange
	p := NewPagination(1, 10, "", OrderAsc, "")

	// Act
	totalPages := p.GetTotalPages(95)

	// Assert
	assert.Equal(suite.T(), 10, totalPages)
}

func (suite *PaginationTestSuite) TestGetTotalPages_LessThanOnePage() {
	// Arrange
	p := NewPagination(1, 10, "", OrderAsc, "")

	// Act
	totalPages := p.GetTotalPages(5)

	// Assert
	assert.Equal(suite.T(), 1, totalPages)
}

func (suite *PaginationTestSuite) TestGetTotalPages_ZeroItems() {
	// Arrange
	p := NewPagination(1, 10, "", OrderAsc, "")

	// Act
	totalPages := p.GetTotalPages(0)

	// Assert
	assert.Equal(suite.T(), 1, totalPages)
}

func (suite *PaginationTestSuite) TestGetTotalPages_OneItem() {
	// Arrange
	p := NewPagination(1, 10, "", OrderAsc, "")

	// Act
	totalPages := p.GetTotalPages(1)

	// Assert
	assert.Equal(suite.T(), 1, totalPages)
}

func (suite *PaginationTestSuite) TestGetTotalPages_WithCustomPageSize() {
	// Arrange
	p := NewPagination(1, 25, "", OrderAsc, "")

	// Act
	totalPages := p.GetTotalPages(100)

	// Assert
	assert.Equal(suite.T(), 4, totalPages)
}

func (suite *PaginationTestSuite) TestGetTotalPages_WithDefaultPageSize() {
	// Arrange - pageSize 0 defaults to 10
	p := NewPagination(1, 0, "", OrderAsc, "")

	// Act
	totalPages := p.GetTotalPages(45)

	// Assert
	assert.Equal(suite.T(), 5, totalPages)
}
