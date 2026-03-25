package handlers

import (
	"encoding/json"

	"github.com/manuel/shopware-testenv-platform/api/internal/registry"
	"gorm.io/datatypes"
)

func mergeRegistryAndDB(reg []registry.MetadataItem, dbJSON datatypes.JSON) []registry.MetadataItem {
	var dbItems []registry.MetadataItem
	if len(dbJSON) > 0 {
		_ = json.Unmarshal(dbJSON, &dbItems)
	}

	dbMap := make(map[string]registry.MetadataItem)
	for _, item := range dbItems {
		dbMap[item.Key] = item
	}

	var merged []registry.MetadataItem
	for _, item := range reg {
		if dbItem, ok := dbMap[item.Key]; ok {
			if dbItem.Value != "" {
				item.Value = dbItem.Value
			}
			delete(dbMap, item.Key)
		}
		merged = append(merged, item)
	}
	for _, item := range dbItems {
		if _, ok := dbMap[item.Key]; ok {
			merged = append(merged, item)
		}
	}

	return merged
}
