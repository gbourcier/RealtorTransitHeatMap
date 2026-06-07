import { api } from './client'

export interface SavedFilterDefinition {
  buildingTypes: number[]
  maxPrice: number | null
  maxCommuteSec: number | null
  newWithinDays: number | null
  minBedrooms: number | null
  minBathrooms: number | null
  minInteriorAreaSqft: number | null
  favoritesOnly: boolean
  includeExpired: boolean
}

export interface SavedFilter extends SavedFilterDefinition {
  id: string
  name: string
  createdAt: number
  updatedAt: number
}

export interface SaveFilterBody extends SavedFilterDefinition {
  name: string
}

export async function listSavedFilters(): Promise<SavedFilter[]> {
  const { data } = await api.get<SavedFilter[]>('/api/saved-filters')
  return data
}

export async function createSavedFilter(body: SaveFilterBody): Promise<SavedFilter> {
  const { data } = await api.post<SavedFilter>('/api/saved-filters', body)
  return data
}

export async function updateSavedFilter(id: string, body: SaveFilterBody): Promise<SavedFilter> {
  const { data } = await api.patch<SavedFilter>(`/api/saved-filters/${id}`, body)
  return data
}

export async function deleteSavedFilter(id: string): Promise<void> {
  await api.delete(`/api/saved-filters/${id}`)
}
