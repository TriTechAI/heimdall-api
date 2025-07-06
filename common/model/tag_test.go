package model

import (
	"testing"
	"time"

	"github.com/heimdall-api/common/constants"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTagModel(t *testing.T) {
	Convey("TagModel测试", t, func() {
		
		Convey("TagModel Validate方法", func() {
			
			Convey("有效的标签数据应该通过验证", func() {
				tag := &TagModel{
					Name:            "Go语言",
					Slug:            "go-language", 
					Description:     "Go编程语言相关文章",
					Color:           "#007d9c",
					FeaturedImage:   "https://example.com/go.png",
					MetaTitle:       "Go编程 - 技术博客",
					MetaDescription: "Go语言编程技术文章",
					Visibility:      constants.TagVisibilityPublic,
				}
				
				err := tag.Validate()
				So(err, ShouldBeNil)
			})
			
			Convey("标签名称验证", func() {
				Convey("空名称应该失败", func() {
					tag := &TagModel{
						Name: "",
						Slug: "test",
						Visibility: constants.TagVisibilityPublic,
					}
					err := tag.Validate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "tag name cannot be empty")
				})
				
				Convey("名称过短应该失败", func() {
					tag := &TagModel{
						Name: "",
						Slug: "test",
						Visibility: constants.TagVisibilityPublic,
					}
					err := tag.Validate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "tag name cannot be empty")
				})
				
				Convey("名称过长应该失败", func() {
					longName := string(make([]rune, constants.TagNameMaxLength+1))
					for i := range longName {
						longName = longName[:i] + "a" + longName[i+1:]
					}
					tag := &TagModel{
						Name: longName,
						Slug: "test",
						Visibility: constants.TagVisibilityPublic,
					}
					err := tag.Validate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "tag name cannot exceed")
				})
			})
			
			Convey("标签slug验证", func() {
				Convey("空slug应该失败", func() {
					tag := &TagModel{
						Name: "Test",
						Slug: "",
						Visibility: constants.TagVisibilityPublic,
					}
					err := tag.Validate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "tag slug cannot be empty")
				})
				
				Convey("slug过长应该失败", func() {
					longSlug := string(make([]byte, constants.TagSlugMaxLength+1))
					for i := range longSlug {
						longSlug = longSlug[:i] + "a" + longSlug[i+1:]
					}
					tag := &TagModel{
						Name: "Test",
						Slug: longSlug,
						Visibility: constants.TagVisibilityPublic,
					}
					err := tag.Validate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "tag slug cannot exceed")
				})
				
				Convey("无效的slug格式应该失败", func() {
					invalidSlugs := []string{
						"Test Slug",      // 包含空格
						"test_slug",      // 包含下划线
						"Test-Slug",      // 包含大写字母
						"test--slug",     // 连续连字符
						"-test-slug",     // 以连字符开头
						"test-slug-",     // 以连字符结尾
						"test@slug",      // 包含特殊字符
						"测试-slug",       // 包含非ASCII字符
					}
					
					for _, invalidSlug := range invalidSlugs {
						tag := &TagModel{
							Name: "Test",
							Slug: invalidSlug,
							Visibility: constants.TagVisibilityPublic,
						}
						err := tag.Validate()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldContainSubstring, "tag slug must contain only lowercase letters")
					}
				})
				
				Convey("有效的slug格式应该通过", func() {
					validSlugs := []string{
						"test",
						"test123",
						"test-slug",
						"test-slug-with-numbers123",
						"a",
						"123",
					}
					
					for _, validSlug := range validSlugs {
						tag := &TagModel{
							Name: "Test",
							Slug: validSlug,
							Visibility: constants.TagVisibilityPublic,
						}
						err := tag.Validate()
						So(err, ShouldBeNil)
					}
				})
			})
			
			Convey("描述验证", func() {
				Convey("空描述应该通过", func() {
					tag := &TagModel{
						Name: "Test",
						Slug: "test",
						Description: "",
						Visibility: constants.TagVisibilityPublic,
					}
					err := tag.Validate()
					So(err, ShouldBeNil)
				})
				
				Convey("描述过长应该失败", func() {
					longDesc := string(make([]rune, constants.TagDescriptionMaxLength+1))
					for i := range longDesc {
						longDesc = longDesc[:i] + "a" + longDesc[i+1:]
					}
					tag := &TagModel{
						Name: "Test",
						Slug: "test",
						Description: longDesc,
						Visibility: constants.TagVisibilityPublic,
					}
					err := tag.Validate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "tag description cannot exceed")
				})
			})
			
			Convey("颜色验证", func() {
				Convey("空颜色应该通过", func() {
					tag := &TagModel{
						Name: "Test",
						Slug: "test",
						Color: "",
						Visibility: constants.TagVisibilityPublic,
					}
					err := tag.Validate()
					So(err, ShouldBeNil)
				})
				
				Convey("有效的hex颜色应该通过", func() {
					validColors := []string{
						"#000000",
						"#FFFFFF",
						"#FF5733",
						"#007d9c",
						"#123abc",
						"#ff5733", // 小写字母也应该有效
					}
					
					for _, validColor := range validColors {
						tag := &TagModel{
							Name: "Test",
							Slug: "test",
							Color: validColor,
							Visibility: constants.TagVisibilityPublic,
						}
						err := tag.Validate()
						So(err, ShouldBeNil)
					}
				})
				
				Convey("无效的颜色格式应该失败", func() {
					invalidColors := []string{
						"red",         // 颜色名称
						"#FF573",      // 长度不足
						"#FF57333",    // 长度过长
						"FF5733",      // 缺少#
						"#GG5733",     // 无效字符
						"rgb(255,87,51)", // RGB格式
					}
					
					for _, invalidColor := range invalidColors {
						tag := &TagModel{
							Name: "Test",
							Slug: "test",
							Color: invalidColor,
							Visibility: constants.TagVisibilityPublic,
						}
						err := tag.Validate()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldContainSubstring, "tag color must be a valid hex color")
					}
				})
			})
			
			Convey("特色图片验证", func() {
				Convey("空图片URL应该通过", func() {
					tag := &TagModel{
						Name: "Test",
						Slug: "test",
						FeaturedImage: "",
						Visibility: constants.TagVisibilityPublic,
					}
					err := tag.Validate()
					So(err, ShouldBeNil)
				})
				
				Convey("图片URL过长应该失败", func() {
					longURL := "https://example.com/" + string(make([]byte, constants.TagFeaturedImageMaxLength))
					tag := &TagModel{
						Name: "Test",
						Slug: "test",
						FeaturedImage: longURL,
						Visibility: constants.TagVisibilityPublic,
					}
					err := tag.Validate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "featured image URL cannot exceed")
				})
			})
			
			Convey("SEO字段验证", func() {
				Convey("MetaTitle过长应该失败", func() {
					longTitle := string(make([]rune, constants.TagMetaTitleMaxLength+1))
					for i := range longTitle {
						longTitle = longTitle[:i] + "a" + longTitle[i+1:]
					}
					tag := &TagModel{
						Name: "Test",
						Slug: "test",
						MetaTitle: longTitle,
						Visibility: constants.TagVisibilityPublic,
					}
					err := tag.Validate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "meta title cannot exceed")
				})
				
				Convey("MetaDescription过长应该失败", func() {
					longDesc := string(make([]rune, constants.TagMetaDescMaxLength+1))
					for i := range longDesc {
						longDesc = longDesc[:i] + "a" + longDesc[i+1:]
					}
					tag := &TagModel{
						Name: "Test",
						Slug: "test",
						MetaDescription: longDesc,
						Visibility: constants.TagVisibilityPublic,
					}
					err := tag.Validate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "meta description cannot exceed")
				})
			})
			
			Convey("可见性验证", func() {
				Convey("空可见性应该通过", func() {
					tag := &TagModel{
						Name: "Test",
						Slug: "test",
						Visibility: "",
					}
					err := tag.Validate()
					So(err, ShouldBeNil)
				})
				
				Convey("有效的可见性应该通过", func() {
					validVisibilities := []string{
						constants.TagVisibilityPublic,
						constants.TagVisibilityInternal,
					}
					
					for _, visibility := range validVisibilities {
						tag := &TagModel{
							Name: "Test",
							Slug: "test",
							Visibility: visibility,
						}
						err := tag.Validate()
						So(err, ShouldBeNil)
					}
				})
				
				Convey("无效的可见性应该失败", func() {
					tag := &TagModel{
						Name: "Test",
						Slug: "test",
						Visibility: "invalid",
					}
					err := tag.Validate()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "invalid visibility")
				})
			})
		})
		
		Convey("TagModel 状态检查方法", func() {
			Convey("IsPublic方法", func() {
				tag := &TagModel{Visibility: constants.TagVisibilityPublic}
				So(tag.IsPublic(), ShouldBeTrue)
				
				tag.Visibility = constants.TagVisibilityInternal
				So(tag.IsPublic(), ShouldBeFalse)
			})
			
			Convey("IsInternal方法", func() {
				tag := &TagModel{Visibility: constants.TagVisibilityInternal}
				So(tag.IsInternal(), ShouldBeTrue)
				
				tag.Visibility = constants.TagVisibilityPublic
				So(tag.IsInternal(), ShouldBeFalse)
			})
		})
		
		Convey("TagModel 默认值设置方法", func() {
			Convey("SetDefaultVisibility方法", func() {
				tag := &TagModel{}
				tag.SetDefaultVisibility()
				So(tag.Visibility, ShouldEqual, constants.TagVisibilityPublic)
				
				// 已有值不应被覆盖
				tag.Visibility = constants.TagVisibilityInternal
				tag.SetDefaultVisibility()
				So(tag.Visibility, ShouldEqual, constants.TagVisibilityInternal)
			})
			
			Convey("SetDefaultColor方法", func() {
				tag := &TagModel{}
				tag.SetDefaultColor()
				So(tag.Color, ShouldEqual, constants.TagDefaultColor)
				
				// 已有值不应被覆盖
				tag.Color = "#FF5733"
				tag.SetDefaultColor()
				So(tag.Color, ShouldEqual, "#FF5733")
			})
		})
		
		Convey("TagModel 数据准备方法", func() {
			Convey("PrepareForCreation方法", func() {
				tag := &TagModel{
					Name: "Test Tag",
					Slug: "test-tag",
				}
				
				beforeTime := time.Now()
				tag.PrepareForCreation()
				afterTime := time.Now()
				
				So(tag.CreatedAt, ShouldHappenOnOrAfter, beforeTime)
				So(tag.CreatedAt, ShouldHappenOnOrBefore, afterTime)
				So(tag.UpdatedAt, ShouldHappenOnOrAfter, beforeTime)
				So(tag.UpdatedAt, ShouldHappenOnOrBefore, afterTime)
				So(tag.PostCount, ShouldEqual, 0)
				So(tag.Visibility, ShouldEqual, constants.TagVisibilityPublic)
				So(tag.Color, ShouldEqual, constants.TagDefaultColor)
			})
			
			Convey("PrepareForUpdate方法", func() {
				tag := &TagModel{
					CreatedAt: time.Now().Add(-time.Hour),
					UpdatedAt: time.Now().Add(-time.Hour),
				}
				oldCreatedAt := tag.CreatedAt
				
				beforeTime := time.Now()
				tag.PrepareForUpdate()
				afterTime := time.Now()
				
				So(tag.CreatedAt, ShouldEqual, oldCreatedAt) // 创建时间不变
				So(tag.UpdatedAt, ShouldHappenOnOrAfter, beforeTime)
				So(tag.UpdatedAt, ShouldHappenOnOrBefore, afterTime)
			})
		})
		
		Convey("TagModel Slug生成方法", func() {
			Convey("GenerateSlugFromName方法", func() {
				Convey("从英文名称生成slug", func() {
					tag := &TagModel{Name: "Go Programming"}
					tag.GenerateSlugFromName()
					So(tag.Slug, ShouldEqual, "go-programming")
				})
				
				Convey("从中文名称生成slug", func() {
					tag := &TagModel{Name: "Go语言编程"}
					tag.GenerateSlugFromName()
					So(tag.Slug, ShouldEqual, "go")
				})
				
				Convey("处理特殊字符", func() {
					tag := &TagModel{Name: "C++ Programming!@#"}
					tag.GenerateSlugFromName()
					So(tag.Slug, ShouldEqual, "c-programming")
				})
				
				Convey("处理长名称", func() {
					longName := "This is a very long tag name that exceeds the maximum slug length limit"
					tag := &TagModel{Name: longName}
					tag.GenerateSlugFromName()
					So(len(tag.Slug), ShouldBeLessThanOrEqualTo, constants.TagSlugMaxLength)
					So(tag.Slug, ShouldNotEndWith, "-")
				})
				
				Convey("已有slug不应被覆盖", func() {
					tag := &TagModel{
						Name: "Go Programming",
						Slug: "existing-slug",
					}
					tag.GenerateSlugFromName()
					So(tag.Slug, ShouldEqual, "existing-slug")
				})
			})
		})
		
		Convey("TagModel 响应转换方法", func() {
			Convey("ToPublicResponse方法", func() {
				tag := &TagModel{
					ID:              primitive.NewObjectID(),
					Name:            "Go语言",
					Slug:            "go-language",
					Description:     "Go编程语言",
					Color:           "#007d9c",
					FeaturedImage:   "https://example.com/go.png",
					MetaTitle:       "Go编程",
					MetaDescription: "Go语言描述",
					PostCount:       25,
					Visibility:      constants.TagVisibilityPublic,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}
				
				response := tag.ToPublicResponse()
				
				So(response["id"], ShouldEqual, tag.ID.Hex())
				So(response["name"], ShouldEqual, tag.Name)
				So(response["slug"], ShouldEqual, tag.Slug)
				So(response["description"], ShouldEqual, tag.Description)
				So(response["color"], ShouldEqual, tag.Color)
				So(response["postCount"], ShouldEqual, tag.PostCount)
				So(response["createdAt"], ShouldEqual, tag.CreatedAt)
				
				// 敏感信息不应包含在公开响应中
				So(response["featuredImage"], ShouldBeNil)
				So(response["metaTitle"], ShouldBeNil)
				So(response["metaDescription"], ShouldBeNil)
				So(response["visibility"], ShouldBeNil)
				So(response["updatedAt"], ShouldBeNil)
			})
			
			Convey("ToAdminResponse方法", func() {
				tag := &TagModel{
					ID:              primitive.NewObjectID(),
					Name:            "Go语言",
					Slug:            "go-language",
					Description:     "Go编程语言",
					Color:           "#007d9c",
					FeaturedImage:   "https://example.com/go.png",
					MetaTitle:       "Go编程",
					MetaDescription: "Go语言描述",
					PostCount:       25,
					Visibility:      constants.TagVisibilityPublic,
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
				}
				
				response := tag.ToAdminResponse()
				
				// 管理员响应包含所有字段
				So(response["id"], ShouldEqual, tag.ID.Hex())
				So(response["name"], ShouldEqual, tag.Name)
				So(response["slug"], ShouldEqual, tag.Slug)
				So(response["description"], ShouldEqual, tag.Description)
				So(response["color"], ShouldEqual, tag.Color)
				So(response["featuredImage"], ShouldEqual, tag.FeaturedImage)
				So(response["metaTitle"], ShouldEqual, tag.MetaTitle)
				So(response["metaDescription"], ShouldEqual, tag.MetaDescription)
				So(response["postCount"], ShouldEqual, tag.PostCount)
				So(response["visibility"], ShouldEqual, tag.Visibility)
				So(response["createdAt"], ShouldEqual, tag.CreatedAt)
				So(response["updatedAt"], ShouldEqual, tag.UpdatedAt)
			})
		})
	})
}

func TestGenerateSlugFromText(t *testing.T) {
	Convey("generateSlugFromText函数测试", t, func() {
		
		Convey("基本功能", func() {
			testCases := []struct {
				input    string
				expected string
			}{
				{"Go Programming", "go-programming"},
				{"JavaScript & TypeScript", "javascript-typescript"},
				{"React.js Development", "reactjs-development"},
				{"  Trimmed Spaces  ", "trimmed-spaces"},
				{"Multiple    Spaces", "multiple-spaces"},
				{"UPPERCASE TEXT", "uppercase-text"},
				{"Mixed-Case_Text", "mixed-case-text"},
				{"123 Numbers 456", "123-numbers-456"},
				{"Special!@#$%Characters", "specialcharacters"},
			}
			
			for _, tc := range testCases {
				result := generateSlugFromText(tc.input)
				So(result, ShouldEqual, tc.expected)
			}
		})
		
		Convey("边界情况", func() {
			Convey("空字符串", func() {
				result := generateSlugFromText("")
				So(result, ShouldEqual, "")
			})
			
			Convey("纯特殊字符", func() {
				result := generateSlugFromText("!@#$%^&*()")
				So(result, ShouldEqual, "")
			})
			
			Convey("连字符处理", func() {
				result := generateSlugFromText("---test---")
				So(result, ShouldEqual, "test")
			})
			
			Convey("长度限制", func() {
				longText := string(make([]rune, constants.TagSlugMaxLength+10))
				for i := range longText {
					longText = longText[:i] + "a" + longText[i+1:]
				}
				result := generateSlugFromText(longText)
				So(len(result), ShouldBeLessThanOrEqualTo, constants.TagSlugMaxLength)
				So(result, ShouldNotEndWith, "-")
			})
		})
	})
}