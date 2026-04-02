package handlers

import (
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
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

func (h *SandboxHandler) enrichSandboxResponses(sandboxes []models.Sandbox) []dto.SandboxResponse {
	type cachedEntry struct {
		metadata []registry.MetadataItem
		ssh      *registry.SSHEntry
	}
	cache := make(map[uuid.UUID]*cachedEntry)

	for idx := range sandboxes {
		sb := &sandboxes[idx]

		entry, ok := cache[sb.ImageID]
		if !ok {
			img, err := h.images.FindByID(sb.ImageID)
			if err != nil {
				slog.Warn("enrich sandbox: image not found", "image_id", sb.ImageID)
				cache[sb.ImageID] = nil
				continue
			}
			regEntry := h.resolver.ResolveEntry(img.RegistryName())
			if regEntry != nil {
				meta := mergeRegistryAndDB(regEntry.Metadata, img.Metadata)
				entry = &cachedEntry{metadata: meta, ssh: regEntry.SSH}
			}
			cache[sb.ImageID] = entry
		}

		if entry == nil {
			continue
		}

		if len(entry.metadata) > 0 {
			var values map[string]string
			if len(sb.Metadata) > 0 {
				_ = json.Unmarshal(sb.Metadata, &values)
			}
			enriched := make([]registry.MetadataItem, len(entry.metadata))
			copy(enriched, entry.metadata)
			for j := range enriched {
				if v, exists := values[enriched[j].Key]; exists {
					enriched[j].Value = v
				}
			}
			data, _ := json.Marshal(enriched)
			sb.Metadata = datatypes.JSON(data)
		}
	}

	responses := toSandboxResponses(sandboxes)

	if h.sshCfg.Enabled {
		for idx := range sandboxes {
			entry := cache[sandboxes[idx].ImageID]
			if entry != nil {
				responses[idx].SSH = buildSSHInfo(&sandboxes[idx], h.sshCfg, entry.ssh)
			}
		}
	}

	return responses
}

func (h *SandboxHandler) enrichSandboxResponse(sandbox *models.Sandbox) dto.SandboxResponse {
	responses := h.enrichSandboxResponses([]models.Sandbox{*sandbox})
	return responses[0]
}
