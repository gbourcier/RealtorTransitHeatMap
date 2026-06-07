import { defineStore } from 'pinia'
import { shallowRef, ref, computed } from 'vue'
import { login as apiLogin, logout as apiLogout, me as apiMe } from '../api/auth'
import type { CurrentUser } from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = shallowRef<CurrentUser | null>(null)
  const defaultFilterId = ref<string | null>(null)
  const initialized = shallowRef(false)

  const isLoggedIn = computed(() => user.value !== null)
  const isAdmin = computed(() => user.value?.role === 'admin')

  function applyUser(u: CurrentUser | null) {
    user.value = u
    defaultFilterId.value = u?.preferences.defaultFilterId ?? null
  }

  async function fetchMe() {
    try {
      applyUser(await apiMe())
    } catch {
      applyUser(null)
    } finally {
      initialized.value = true
    }
  }

  async function login(username: string, password: string) {
    applyUser(await apiLogin(username, password))
    initialized.value = true
  }

  async function logout() {
    try {
      await apiLogout()
    } finally {
      applyUser(null)
    }
  }

  function setDefaultFilterId(id: string | null) {
    defaultFilterId.value = id
  }

  return { user, defaultFilterId, initialized, isLoggedIn, isAdmin, fetchMe, login, logout, setDefaultFilterId }
})
