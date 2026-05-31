import { defineStore } from 'pinia'
import { shallowRef, computed } from 'vue'
import { login as apiLogin, logout as apiLogout, me as apiMe } from '../api/auth'
import type { User } from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = shallowRef<User | null>(null)
  const initialized = shallowRef(false)

  const isLoggedIn = computed(() => user.value !== null)
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function fetchMe() {
    try {
      user.value = await apiMe()
    } catch {
      user.value = null
    } finally {
      initialized.value = true
    }
  }

  async function login(username: string, password: string) {
    user.value = await apiLogin(username, password)
    initialized.value = true
  }

  async function logout() {
    try {
      await apiLogout()
    } finally {
      user.value = null
    }
  }

  return { user, initialized, isLoggedIn, isAdmin, fetchMe, login, logout }
})
