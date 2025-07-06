package main

import (
	"fmt"
	"reflect"

	adminSvc "github.com/heimdall-api/admin-api/admin/internal/svc"
	publicSvc "github.com/heimdall-api/public-api/public/internal/svc"
	"github.com/heimdall-api/common/dao"
)

func main() {
	fmt.Println("🔍 Verifying TagDAO Service Context Integration...")

	// Check admin ServiceContext has TagDAO field
	adminType := reflect.TypeOf(adminSvc.ServiceContext{})
	adminTagDAOField, adminHasTagDAO := adminType.FieldByName("TagDAO")
	if adminHasTagDAO {
		fmt.Printf("✅ Admin ServiceContext has TagDAO field of type: %s\n", adminTagDAOField.Type)
		if adminTagDAOField.Type == reflect.TypeOf(&dao.TagDAO{}) {
			fmt.Println("✅ Admin TagDAO field type is correct: *dao.TagDAO")
		}
	} else {
		fmt.Println("❌ Admin ServiceContext missing TagDAO field")
	}

	// Check public ServiceContext has TagDAO field  
	publicType := reflect.TypeOf(publicSvc.ServiceContext{})
	publicTagDAOField, publicHasTagDAO := publicType.FieldByName("TagDAO")
	if publicHasTagDAO {
		fmt.Printf("✅ Public ServiceContext has TagDAO field of type: %s\n", publicTagDAOField.Type)
		if publicTagDAOField.Type == reflect.TypeOf(&dao.TagDAO{}) {
			fmt.Println("✅ Public TagDAO field type is correct: *dao.TagDAO")
		}
	} else {
		fmt.Println("❌ Public ServiceContext missing TagDAO field")
	}

	if adminHasTagDAO && publicHasTagDAO {
		fmt.Println("🎉 TagDAO Service Context Integration Verification PASSED!")
	} else {
		fmt.Println("❌ TagDAO Service Context Integration Verification FAILED!")
	}
}