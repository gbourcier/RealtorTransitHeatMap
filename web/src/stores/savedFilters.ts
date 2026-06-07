import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  listSavedFilters,
  createSavedFilter,
  updateSavedFilter,
  deleteSavedFilter,
  type SavedFilter,
  type SavedFilterDefinition,
} from '../api/savedFilters'
import { setDefaultFilter } from '../api/preferences'
import { useAuthStore } from './auth'

export const useSavedFiltersStore = defineStore('savedFilters', () => {
  const list = ref<SavedFilter[]>([])
  const initialized = ref(false)

  const authStore = useAuthStore()

  const defaultFilter = computed(
    () => list.value.find((f) => f.id === authStore.defaultFilterId) ?? null,
  )

  async function fetchList() {
    list.value = await listSavedFilters()
    initialized.value = true
  }

  async function create(name: string, def: SavedFilterDefinition) {
    const created = await createSavedFilter({ name, ...def })
    list.value = [...list.value, created]
    return created
  }

  async function update(id: string, name: string, def: SavedFilterDefinition) {
    const updated = await updateSavedFilter(id, { name, ...def })
    list.value = list.value.map((f) => (f.id === id ? updated : f))
    return updated
  }

  async function remove(id: string) {
    await deleteSavedFilter(id)
    list.value = list.value.filter((f) => f.id !== id)
    if (authStore.defaultFilterId === id) {
      authStore.setDefaultFilterId(null)
    }
  }

  async function setDefault(id: string | null) {
    await setDefaultFilter(id)
    authStore.setDefaultFilterId(id)
  }

  function reset() {
    list.value = []
    initialized.value = false
  }

  return { list, initialized, defaultFilter, fetchList, create, update, remove, setDefault, reset }
})
