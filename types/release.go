package types

// {
// 	"announcement": true,
// 	"tags": [
// 		"security"
// 	],
// 	"date": "21 Nov 2024",
// 	"source": [
// 		{
// 			"filename": "php-8.4.1.tar.gz",
// 			"name": "PHP 8.4.1 (tar.gz)",
// 			"sha256": "c3d1ce4157463ea43004289c01172deb54ce9c5894d8722f4e805461bf9feaec",
// 			"date": "21 Nov 2024"
// 		},
// 		{
// 			"filename": "php-8.4.1.tar.bz2",
// 			"name": "PHP 8.4.1 (tar.bz2)",
// 			"sha256": "ef8a270118ed128b765fc31f198c7f4650c8171411b0f6a3a1a3aba11fcacc23",
// 			"date": "21 Nov 2024"
// 		},
// 		{
// 			"filename": "php-8.4.1.tar.xz",
// 			"name": "PHP 8.4.1 (tar.xz)",
// 			"sha256": "94c8a4fd419d45748951fa6d73bd55f6bdf0adaefb8814880a67baa66027311f",
// 			"date": "21 Nov 2024"
// 		}
// 	],
// 	"version": "8.4.1",
// 	"supported_versions": [
// 		"8.1",
// 		"8.2",
// 		"8.3",
// 		"8.4"
// 	]
// }

type Source struct {
	Filename string `json:"filename"`
	Name     string `json:"name"`
	Sha256   string `json:"sha256"`
	Date     string `json:"date"`
}

type Release struct {
	Source            []Source `json:"source"`
	SupportedVersions []string `json:"supported_versions"`
	Version           string   `json:"version"`
}

type ReleaseData map[string]Release

type ReleaseDataWrapper struct {
	Data ReleaseData
}

func (rw ReleaseDataWrapper) ToSlice() []Release {
	var releases []Release
	for _, release := range rw.Data {
		releases = append(releases, release)
	}
	return releases
}

func (r *ReleaseData) Len()  {}
func (r *ReleaseData) Less() {}
func (r *ReleaseData) Swap() {}
