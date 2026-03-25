import * as icons from 'lucide-vue-next'

import type { Component } from 'vue'

const iconMap = icons as unknown as Record<string, Component>
const resolveCache = new Map<string, Component | undefined>()

export function resolveIcon(name?: string): Component | undefined {
  if (!name) return undefined
  if (resolveCache.has(name)) return resolveCache.get(name)
  const pascal = name.replace(/(^|-)(\w)/g, (_, __, c: string) => c.toUpperCase())
  const icon = iconMap[pascal]
  resolveCache.set(name, icon)
  return icon
}

let cachedNames: string[] | null = null

export function getIconNames(): string[] {
  if (cachedNames) return cachedNames
  cachedNames = Object.keys(iconMap)
    .filter((k) => k !== 'default' && k !== 'createIcons' && k[0] === k[0].toUpperCase())
    .map((k) => k.replace(/([a-z])([A-Z])/g, '$1-$2').toLowerCase())
    .sort()
  return cachedNames
}
