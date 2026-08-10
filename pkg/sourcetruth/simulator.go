package sourcetruth

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/khhini/open-knowledge-engine.git/pkg/okf"
)

type Simulator struct {
	RootDir string
}

func NewSimulator(rootDir string) *Simulator {
	if rootDir == "" {
		rootDir = ".source_of_truth"
	}
	return &Simulator{RootDir: rootDir}
}

func (s *Simulator) Inspect(uriStr string) (*Inspection, error) {
	dataPath, metaPath, assetType := ResolveLocalPath(s.RootDir, uriStr)

	inspection := &Inspection{
		URI:       uriStr,
		AssetType: assetType,
		LocalPath: dataPath,
		Exists:    false,
	}

	if metaPath != "" {
		if metaBytes, err := os.ReadFile(metaPath); err == nil {
			var meta map[string]any
			if err := json.Unmarshal(metaBytes, &meta); err == nil {
				inspection.Exists = true

				if name, ok := meta["name"].(string); ok {
					inspection.FileName = name
				}
				if mime, ok := meta["mime_type"].(string); ok {
					inspection.MimeType = mime
				}
				if owner, ok := meta["owner"].(string); ok {
					inspection.Owner = owner
				}
				if summary, ok := meta["document_summary"].(string); ok {
					inspection.Summary = summary
				}
				if pages, ok := meta["page_count"].(float64); ok {
					inspection.PageCount = int(pages)
				}
				if size, ok := meta["size_bytes"].(float64); ok {
					inspection.SizeBytes = int64(size)
				}
				if modStr, ok := meta["last_modified"].(string); ok {
					if t, err := time.Parse(time.RFC3339, modStr); err == nil {
						inspection.LastModified = t
					}
				}
				// If metadata contains columns
				if colsRaw, ok := meta["columns"].([]any); ok {
					for _, c := range colsRaw {
						if colName, ok := c.(string); ok {
							inspection.Columns = append(inspection.Columns, Column{Name: colName, Type: "STRING"})
						} else if colMap, ok := c.(map[string]any); ok {
							col := Column{
								Name: fmt.Sprintf("%v", colMap["name"]),
								Type: fmt.Sprintf("%v", colMap["type"]),
							}
							if desc, ok := colMap["description"].(string); ok {
								col.Description = desc
							}
							inspection.Columns = append(inspection.Columns, col)
						}
					}
				}
			}
		}
	}

	if dataBytes, err := os.ReadFile(dataPath); err == nil {
		inspection.Exists = true
		content := string(dataBytes)
		inspection.RawContent = content

		if len(content) > 1000 {
			inspection.TextSnippet = content[:1000] + "\n... [Truncated preview]"
		} else {
			inspection.TextSnippet = content
		}

		if inspection.SizeBytes == 0 {
			inspection.SizeBytes = int64(len(dataBytes))
		}
	}

	if !inspection.Exists {
		inspection.ErrorMessage = fmt.Sprintf("Simulated not found at path: %s", dataPath)
	}

	return inspection, nil
}

func (s *Simulator) ValidateSources(sources []okf.Source) ([]SourceStatus, error) {
	var results []SourceStatus
	for _, src := range sources {
		insp, _ := s.Inspect(src.Resource)
		results = append(results, SourceStatus{
			SourceID:     src.ID,
			URI:          src.Resource,
			AssetType:    insp.AssetType,
			Exists:       insp.Exists,
			LocalPath:    insp.LocalPath,
			LastModified: src.LastModified,
		})
	}

	return results, nil
}
