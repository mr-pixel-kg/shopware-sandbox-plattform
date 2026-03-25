import type { MetadataItem } from '@/types'

export interface MetadataRow {
  key: string
  label: string
  value: string
  fromRegistry: boolean
  icon: string
  show: string
  condition: string
  size: string
}

export function newMetadataRow(): MetadataRow {
  return {
    key: '',
    label: '',
    value: '',
    fromRegistry: false,
    icon: '',
    show: 'sandbox',
    condition: 'always',
    size: 'default',
  }
}

export function rowToMetadataItem(row: MetadataRow, type: MetadataItem['type']): MetadataItem {
  return {
    key: row.key || row.label.toLowerCase().replace(/\s+/g, '_'),
    label: row.label,
    type,
    value: row.value,
    icon: row.icon || undefined,
    show: (row.show as MetadataItem['show']) || undefined,
    condition: (row.condition as MetadataItem['condition']) || undefined,
    size: (row.size as MetadataItem['size']) || undefined,
  }
}

export function metadataItemToRow(m: MetadataItem, fromRegistry: boolean): MetadataRow {
  return {
    key: m.key,
    label: m.label,
    value: m.value ?? '',
    fromRegistry,
    icon: m.icon ?? '',
    show: m.show ?? 'sandbox',
    condition: m.condition ?? 'always',
    size: m.size ?? 'default',
  }
}

export function collectMetadata(
  fieldRows: MetadataRow[],
  actionRows: MetadataRow[],
): MetadataItem[] {
  const metadata: MetadataItem[] = []

  for (const row of fieldRows) {
    if (!row.label) continue
    metadata.push(rowToMetadataItem(row, 'field'))
  }

  for (const row of actionRows) {
    if (row.fromRegistry || !row.label) continue
    metadata.push(rowToMetadataItem(row, 'action'))
  }

  return metadata
}
