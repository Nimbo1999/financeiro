package messaging

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RetryTestSuite struct {
	suite.Suite
}

func (suite *RetryTestSuite) TestGetRetryCount_NoHeaders() {
	delivery := amqp.Delivery{}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_EmptyHeaders() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_NoXDeathHeader() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"other-header": "value",
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_XDeathEmptyArray() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{},
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_XDeathInvalidType() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": "invalid-type",
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_XDeathRecordInvalidType() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				"invalid-record",
			},
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_CountMissing() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{
					"reason": "rejected",
				},
			},
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_CountInvalidType() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{
					"count": "not-a-number",
				},
			},
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_CustomHeader0() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			RetryCountHeader: int32(0),
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_CustomHeader3() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			RetryCountHeader: int32(3),
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(3), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_CustomHeader5() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			RetryCountHeader: int32(5),
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(5), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_CustomHeaderTakesPrecedence() {
	// When both custom header and x-death exist, custom header should take precedence
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			RetryCountHeader: int32(3),
			"x-death": []interface{}{
				amqp.Table{
					"count": int64(10),
				},
			},
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(3), count, "Custom header should take precedence over x-death")
}

func (suite *RetryTestSuite) TestGetRetryCount_Count0() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{
					"count": int64(0),
				},
			},
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_Count1() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{
					"count": int64(1),
				},
			},
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_Count3() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{
					"count": int64(3),
				},
			},
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(3), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_Count5() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{
					"count": int64(5),
				},
			},
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(5), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_Count10() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{
					"count": int64(10),
				},
			},
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(10), count)
}

func (suite *RetryTestSuite) TestGetRetryCount_MultipleDeathRecords() {
	// Only the first death record should be used
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{
					"count": int64(3),
				},
				amqp.Table{
					"count": int64(5),
				},
			},
		},
	}
	count := GetRetryCount(delivery)
	assert.Equal(suite.T(), int64(3), count)
}

func (suite *RetryTestSuite) TestShouldRetry_Attempt0() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{"count": int64(0)},
			},
		},
	}
	assert.True(suite.T(), ShouldRetry(delivery))
	assert.False(suite.T(), ShouldMoveToDLQ(delivery))
}

func (suite *RetryTestSuite) TestShouldRetry_Attempt1() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{"count": int64(1)},
			},
		},
	}
	assert.True(suite.T(), ShouldRetry(delivery))
	assert.False(suite.T(), ShouldMoveToDLQ(delivery))
}

func (suite *RetryTestSuite) TestShouldRetry_Attempt2() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{"count": int64(2)},
			},
		},
	}
	assert.True(suite.T(), ShouldRetry(delivery))
	assert.False(suite.T(), ShouldMoveToDLQ(delivery))
}

func (suite *RetryTestSuite) TestShouldRetry_Attempt3() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{"count": int64(3)},
			},
		},
	}
	assert.True(suite.T(), ShouldRetry(delivery))
	assert.False(suite.T(), ShouldMoveToDLQ(delivery))
}

func (suite *RetryTestSuite) TestShouldRetry_Attempt4() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{"count": int64(4)},
			},
		},
	}
	assert.True(suite.T(), ShouldRetry(delivery))
	assert.False(suite.T(), ShouldMoveToDLQ(delivery))
}

func (suite *RetryTestSuite) TestShouldRetry_Attempt5_ShouldNotRetry() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{"count": int64(5)},
			},
		},
	}
	assert.False(suite.T(), ShouldRetry(delivery))
	assert.True(suite.T(), ShouldMoveToDLQ(delivery))
}

func (suite *RetryTestSuite) TestShouldRetry_Attempt10_ShouldNotRetry() {
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{"count": int64(10)},
			},
		},
	}
	assert.False(suite.T(), ShouldRetry(delivery))
	assert.True(suite.T(), ShouldMoveToDLQ(delivery))
}

func (suite *RetryTestSuite) TestShouldRetry_NoHeaders_ShouldRetry() {
	delivery := amqp.Delivery{}
	assert.True(suite.T(), ShouldRetry(delivery))
	assert.False(suite.T(), ShouldMoveToDLQ(delivery))
}

func (suite *RetryTestSuite) TestMaxRetries_Constant() {
	// Verify MaxRetries constant is set to 5 as per specification
	assert.Equal(suite.T(), 5, MaxRetries)
}

func (suite *RetryTestSuite) TestShouldRetry_BoundaryCondition_JustBeforeMax() {
	// Test boundary: MaxRetries - 1 should retry
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{"count": int64(MaxRetries - 1)},
			},
		},
	}
	assert.True(suite.T(), ShouldRetry(delivery))
	assert.False(suite.T(), ShouldMoveToDLQ(delivery))
}

func (suite *RetryTestSuite) TestShouldRetry_BoundaryCondition_ExactlyMax() {
	// Test boundary: exactly MaxRetries should not retry
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{"count": int64(MaxRetries)},
			},
		},
	}
	assert.False(suite.T(), ShouldRetry(delivery))
	assert.True(suite.T(), ShouldMoveToDLQ(delivery))
}

func (suite *RetryTestSuite) TestShouldRetry_BoundaryCondition_JustAfterMax() {
	// Test boundary: MaxRetries + 1 should not retry
	delivery := amqp.Delivery{
		Headers: amqp.Table{
			"x-death": []interface{}{
				amqp.Table{"count": int64(MaxRetries + 1)},
			},
		},
	}
	assert.False(suite.T(), ShouldRetry(delivery))
	assert.True(suite.T(), ShouldMoveToDLQ(delivery))
}

func (suite *RetryTestSuite) TestShouldMoveToDLQ_InverseOfShouldRetry() {
	// Verify that ShouldMoveToDLQ is always the inverse of ShouldRetry
	testCases := []int64{0, 1, 2, 3, 4, 5, 6, 10, 100}

	for _, retryCount := range testCases {
		delivery := amqp.Delivery{
			Headers: amqp.Table{
				"x-death": []interface{}{
					amqp.Table{"count": retryCount},
				},
			},
		}

		shouldRetry := ShouldRetry(delivery)
		shouldMoveToDLQ := ShouldMoveToDLQ(delivery)

		// They should always be opposite
		assert.NotEqual(suite.T(), shouldRetry, shouldMoveToDLQ,
			"ShouldRetry and ShouldMoveToDLQ should be inverse for count=%d", retryCount)
	}
}

func TestRetryTestSuite(t *testing.T) {
	suite.Run(t, new(RetryTestSuite))
}
