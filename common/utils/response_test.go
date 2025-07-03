package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestResponse(t *testing.T) {
	Convey("Test Response Basic Functionality", t, func() {
		resp := &Response{
			Code:      StatusOK,
			Message:   "Operation successful",
			Data:      map[string]string{"test": "data"},
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		Convey("Should create response correctly", func() {
			So(resp, ShouldNotBeNil)
			So(resp.Code, ShouldEqual, StatusOK)
			So(resp.Message, ShouldEqual, "Operation successful")
			So(resp.Data, ShouldNotBeNil)
			So(resp.Timestamp, ShouldNotBeEmpty)
		})
	})
}

func TestErrorResponse(t *testing.T) {
	Convey("Test ErrorResponse Basic Functionality", t, func() {
		errResp := &ErrorResponse{
			Code:    ErrCodeValidationFailed,
			Message: "Invalid input data",
			Details: map[string]interface{}{
				"field": "email",
				"error": "invalid format",
			},
		}

		Convey("Should create error response correctly", func() {
			So(errResp, ShouldNotBeNil)
			So(errResp.Code, ShouldEqual, ErrCodeValidationFailed)
			So(errResp.Message, ShouldEqual, "Invalid input data")
			So(errResp.Details, ShouldNotBeNil)
		})
	})
}

func TestSuccessResponse(t *testing.T) {
	Convey("Test Success Response Functions", t, func() {
		Convey("Success with data should work", func() {
			recorder := httptest.NewRecorder()
			data := map[string]string{"message": "test"}
			Success(recorder, data)

			So(recorder.Code, ShouldEqual, http.StatusOK)
			So(recorder.Header().Get("Content-Type"), ShouldEqual, "application/json; charset=utf-8")

			var response Response
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, StatusOK)
			So(response.Message, ShouldEqual, "Success")
			So(response.Data, ShouldNotBeNil)
		})

		Convey("Success with nil data should work", func() {
			recorder := httptest.NewRecorder()
			Success(recorder, nil)

			So(recorder.Code, ShouldEqual, http.StatusOK)

			var response Response
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, StatusOK)
		})

		Convey("Created should work", func() {
			recorder := httptest.NewRecorder()
			data := map[string]string{"id": "123"}
			Created(recorder, data)

			So(recorder.Code, ShouldEqual, http.StatusCreated)

			var response Response
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, StatusCreated)
			So(response.Message, ShouldEqual, "Created")
		})

		Convey("NoContent should work", func() {
			recorder := httptest.NewRecorder()
			NoContent(recorder)

			So(recorder.Code, ShouldEqual, http.StatusNoContent)
			So(recorder.Body.Len(), ShouldEqual, 0)
		})
	})
}

func TestErrorResponses(t *testing.T) {
	Convey("Test Error Response Functions", t, func() {
		Convey("BadRequest should work", func() {
			recorder := httptest.NewRecorder()
			message := "Invalid request data"
			details := map[string]string{"field": "email"}
			BadRequest(recorder, message, details)

			So(recorder.Code, ShouldEqual, http.StatusBadRequest)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, ErrCodeValidationFailed)
			So(response.Message, ShouldEqual, message)
		})

		Convey("Unauthorized should work", func() {
			recorder := httptest.NewRecorder()
			Unauthorized(recorder, "Custom unauthorized message")

			So(recorder.Code, ShouldEqual, http.StatusUnauthorized)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, ErrCodeUnauthorized)
			So(response.Message, ShouldEqual, "Custom unauthorized message")
		})

		Convey("Unauthorized with empty message should use default", func() {
			recorder := httptest.NewRecorder()
			Unauthorized(recorder, "")

			So(recorder.Code, ShouldEqual, http.StatusUnauthorized)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Message, ShouldEqual, "Authentication required")
		})

		Convey("Forbidden should work", func() {
			recorder := httptest.NewRecorder()
			Forbidden(recorder, "Custom forbidden message")

			So(recorder.Code, ShouldEqual, http.StatusForbidden)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, ErrCodeForbidden)
			So(response.Message, ShouldEqual, "Custom forbidden message")
		})

		Convey("NotFound should work", func() {
			recorder := httptest.NewRecorder()
			NotFound(recorder, "Custom not found message")

			So(recorder.Code, ShouldEqual, http.StatusNotFound)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, ErrCodeNotFound)
			So(response.Message, ShouldEqual, "Custom not found message")
		})

		Convey("InternalError should work", func() {
			recorder := httptest.NewRecorder()
			InternalError(recorder, "Custom internal error")

			So(recorder.Code, ShouldEqual, http.StatusInternalServerError)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, ErrCodeInternalError)
			So(response.Message, ShouldEqual, "Custom internal error")
		})

		Convey("ValidationError should work", func() {
			recorder := httptest.NewRecorder()
			errors := map[string][]string{
				"email": {"invalid format", "required"},
				"name":  {"too short"},
			}
			ValidationError(recorder, errors)

			So(recorder.Code, ShouldEqual, http.StatusBadRequest)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, ErrCodeValidationFailed)
			So(response.Message, ShouldEqual, "Validation failed")
			So(response.Details, ShouldNotBeNil)
		})
	})
}

func TestPaginationResponse(t *testing.T) {
	Convey("Test Pagination Response", t, func() {
		Convey("SuccessWithPagination should work", func() {
			recorder := httptest.NewRecorder()

			data := []map[string]string{
				{"id": "1", "name": "item1"},
				{"id": "2", "name": "item2"},
			}

			SuccessWithPagination(recorder, data, 1, 10, 100)

			So(recorder.Code, ShouldEqual, http.StatusOK)

			var response Response
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)

			So(response.Code, ShouldEqual, StatusOK)
			So(response.Data, ShouldNotBeNil)

			// 检查分页数据结构
			dataMap := response.Data.(map[string]interface{})
			So(dataMap["list"], ShouldNotBeNil)
			So(dataMap["pagination"], ShouldNotBeNil)

			pagination := dataMap["pagination"].(map[string]interface{})
			So(pagination["page"], ShouldEqual, float64(1))
			So(pagination["limit"], ShouldEqual, float64(10))
			So(pagination["total"], ShouldEqual, float64(100))
			So(pagination["totalPages"], ShouldEqual, float64(10))
		})

		Convey("Empty pagination should work", func() {
			recorder := httptest.NewRecorder()

			data := []interface{}{}
			SuccessWithPagination(recorder, data, 1, 10, 0)

			So(recorder.Code, ShouldEqual, http.StatusOK)

			var response Response
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)

			dataMap := response.Data.(map[string]interface{})
			pagination := dataMap["pagination"].(map[string]interface{})
			So(pagination["total"], ShouldEqual, float64(0))
			So(pagination["totalPages"], ShouldEqual, float64(0))
		})
	})
}

func TestCustomResponse(t *testing.T) {
	Convey("Test Custom Response Functions", t, func() {
		Convey("CustomError should work", func() {
			recorder := httptest.NewRecorder()
			customCode := "CUSTOM_ERROR"
			customMessage := "Custom error occurred"
			details := map[string]interface{}{
				"error_type": "business_logic",
				"field":      "amount",
			}

			CustomError(recorder, http.StatusConflict, customCode, customMessage, details)

			So(recorder.Code, ShouldEqual, http.StatusConflict)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, customCode)
			So(response.Message, ShouldEqual, customMessage)
			So(response.Details, ShouldNotBeNil)
		})

		Convey("Error function should work", func() {
			recorder := httptest.NewRecorder()
			Error(recorder, http.StatusTeapot, "TEAPOT_ERROR", "I'm a teapot", nil)

			So(recorder.Code, ShouldEqual, http.StatusTeapot)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, "TEAPOT_ERROR")
			So(response.Message, ShouldEqual, "I'm a teapot")
		})
	})
}

func TestSpecializedErrors(t *testing.T) {
	Convey("Test Specialized Error Functions", t, func() {
		Convey("TokenExpired should work", func() {
			recorder := httptest.NewRecorder()
			TokenExpired(recorder)

			So(recorder.Code, ShouldEqual, http.StatusUnauthorized)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, ErrCodeTokenExpired)
		})

		Convey("TokenInvalid should work", func() {
			recorder := httptest.NewRecorder()
			TokenInvalid(recorder)

			So(recorder.Code, ShouldEqual, http.StatusUnauthorized)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, ErrCodeInvalidToken)
		})

		Convey("UsernameExists should work", func() {
			recorder := httptest.NewRecorder()
			UsernameExists(recorder)

			So(recorder.Code, ShouldEqual, http.StatusConflict)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, ErrCodeUsernameExists)
		})

		Convey("EmailExists should work", func() {
			recorder := httptest.NewRecorder()
			EmailExists(recorder)

			So(recorder.Code, ShouldEqual, http.StatusConflict)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, ErrCodeEmailExists)
		})

		Convey("AccountLocked should work", func() {
			recorder := httptest.NewRecorder()
			AccountLocked(recorder, "Account locked due to too many failed attempts")

			So(recorder.Code, ShouldEqual, http.StatusForbidden)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Code, ShouldEqual, ErrCodeAccountLocked)
		})
	})
}

func TestResponseHelpers(t *testing.T) {
	Convey("Test Response Helper Functions", t, func() {
		Convey("SetSecurityHeaders should work", func() {
			recorder := httptest.NewRecorder()
			SetSecurityHeaders(recorder)

			headers := recorder.Header()
			So(headers.Get("X-Content-Type-Options"), ShouldEqual, "nosniff")
			So(headers.Get("X-Frame-Options"), ShouldEqual, "DENY")
			So(headers.Get("X-XSS-Protection"), ShouldEqual, "1; mode=block")
		})

		Convey("SetCacheHeaders should work", func() {
			recorder := httptest.NewRecorder()
			SetCacheHeaders(recorder, 3600)

			headers := recorder.Header()
			So(headers.Get("Cache-Control"), ShouldContainSubstring, "max-age=3600")
		})

		Convey("SetCORSHeaders should work", func() {
			recorder := httptest.NewRecorder()
			SetCORSHeaders(recorder, "*")

			headers := recorder.Header()
			So(headers.Get("Access-Control-Allow-Origin"), ShouldEqual, "*")
		})

		Convey("HandlePreflight should work", func() {
			recorder := httptest.NewRecorder()
			HandlePreflight(recorder)

			So(recorder.Code, ShouldEqual, http.StatusNoContent)
		})
	})
}

func TestPaginationMetadata(t *testing.T) {
	Convey("Test Pagination Metadata", t, func() {
		Convey("CreatePaginationMetadata should work correctly", func() {
			metadata := CreatePaginationMetadata(2, 10, 25)

			So(metadata.Page, ShouldEqual, 2)
			So(metadata.Limit, ShouldEqual, 10)
			So(metadata.Total, ShouldEqual, 25)
			So(metadata.TotalPages, ShouldEqual, 3)
			So(metadata.HasNext, ShouldBeTrue)
			So(metadata.HasPrev, ShouldBeTrue)
		})

		Convey("CreatePaginationMetadata for first page", func() {
			metadata := CreatePaginationMetadata(1, 10, 25)

			So(metadata.HasNext, ShouldBeTrue)
			So(metadata.HasPrev, ShouldBeFalse)
		})

		Convey("CreatePaginationMetadata for last page", func() {
			metadata := CreatePaginationMetadata(3, 10, 25)

			So(metadata.HasNext, ShouldBeFalse)
			So(metadata.HasPrev, ShouldBeTrue)
		})
	})
}

func TestAPIResponseInterface(t *testing.T) {
	Convey("Test APIResponse Interface", t, func() {
		Convey("NewSuccessResponse should work", func() {
			data := map[string]string{"test": "data"}
			response := NewSuccessResponse(data)

			So(response, ShouldNotBeNil)

			recorder := httptest.NewRecorder()
			response.WriteResponse(recorder)

			So(recorder.Code, ShouldEqual, http.StatusOK)
		})

		Convey("NewErrorResponse should work", func() {
			response := NewErrorResponse(http.StatusBadRequest, "TEST_ERROR", "Test error", nil)

			So(response, ShouldNotBeNil)

			recorder := httptest.NewRecorder()
			response.WriteResponse(recorder)

			So(recorder.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestResponseMiddleware(t *testing.T) {
	Convey("Test Response Middleware", t, func() {
		Convey("NewResponseMiddleware should work", func() {
			middleware := NewResponseMiddleware(true, true, "*")

			So(middleware, ShouldNotBeNil)
			So(middleware.enableSecurity, ShouldBeTrue)
			So(middleware.enableCORS, ShouldBeTrue)
			So(middleware.corsOrigin, ShouldEqual, "*")
		})

		Convey("ResponseMiddleware.Wrap should work", func() {
			middleware := NewResponseMiddleware(true, false, "")
			recorder := httptest.NewRecorder()

			wrappedWriter := middleware.Wrap(recorder)
			So(wrappedWriter, ShouldNotBeNil)
		})
	})
}

func TestResponseEdgeCases(t *testing.T) {
	Convey("Test Edge Cases", t, func() {
		Convey("Very large data should be handled", func() {
			recorder := httptest.NewRecorder()

			// 创建大量数据
			largeData := make(map[string]interface{})
			for i := 0; i < 1000; i++ {
				largeData[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
			}

			Success(recorder, largeData)

			So(recorder.Code, ShouldEqual, http.StatusOK)
			So(recorder.Body.Len(), ShouldBeGreaterThan, 10000)
		})

		Convey("Nil data should not cause panic", func() {
			recorder := httptest.NewRecorder()

			So(func() {
				Success(recorder, nil)
			}, ShouldNotPanic)

			So(recorder.Code, ShouldEqual, http.StatusOK)
		})

		Convey("Empty string message should work", func() {
			recorder := httptest.NewRecorder()
			Unauthorized(recorder, "")

			So(recorder.Code, ShouldEqual, http.StatusUnauthorized)

			var response ErrorResponse
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)
			So(response.Message, ShouldEqual, "Authentication required")
		})

		Convey("Zero pagination values should work", func() {
			recorder := httptest.NewRecorder()
			SuccessWithPagination(recorder, []interface{}{}, 1, 10, 0)

			So(recorder.Code, ShouldEqual, http.StatusOK)

			var response Response
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)

			dataMap := response.Data.(map[string]interface{})
			pagination := dataMap["pagination"].(map[string]interface{})
			So(pagination["total"], ShouldEqual, float64(0))
		})
	})
}

func TestResponseTiming(t *testing.T) {
	Convey("Test Response Timing", t, func() {
		Convey("Response timestamp should be recent", func() {
			recorder := httptest.NewRecorder()
			beforeTime := time.Now().UTC()

			Success(recorder, nil)

			afterTime := time.Now().UTC()

			var response Response
			err := json.Unmarshal(recorder.Body.Bytes(), &response)
			So(err, ShouldBeNil)

			timestamp, err := time.Parse(time.RFC3339, response.Timestamp)
			So(err, ShouldBeNil)
			So(timestamp, ShouldHappenBetween, beforeTime.Add(-time.Second), afterTime.Add(time.Second))
		})
	})
}

func BenchmarkSuccessResponse(b *testing.B) {
	data := map[string]string{"test": "data"}
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		Success(recorder, data)
	}
}

func BenchmarkErrorResponse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		BadRequest(recorder, "Test error message", nil)
	}
}

func BenchmarkPaginationResponse(b *testing.B) {
	data := []map[string]string{
		{"id": "1", "name": "item1"},
		{"id": "2", "name": "item2"},
	}

	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		SuccessWithPagination(recorder, data, 1, 10, 100)
	}
}

func BenchmarkJSONSerialization(b *testing.B) {
	data := map[string]interface{}{
		"string": "test",
		"number": 123,
		"array":  []int{1, 2, 3, 4, 5},
		"object": map[string]string{"nested": "value"},
	}

	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		Success(recorder, data)
	}
}
