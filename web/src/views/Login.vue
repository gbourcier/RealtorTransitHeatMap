<template>
  <v-main class="login-main">
    <v-container class="fill-height" fluid>
      <v-row align="center" justify="center">
        <v-col cols="12" sm="8" md="5" lg="4">
          <v-card elevation="4" class="pa-2">
            <v-card-title class="pt-6 pb-2 text-center text-h6">
              Realtor Transit Heatmap
            </v-card-title>
            <v-card-subtitle class="text-center pb-4">Sign in to continue</v-card-subtitle>

            <v-card-text>
              <v-form @submit.prevent="onSubmit">
                <v-text-field
                  v-model="username"
                  label="Username"
                  prepend-inner-icon="mdi-account-outline"
                  autocomplete="username"
                  variant="outlined"
                  density="comfortable"
                  class="mb-3"
                  :disabled="loading"
                />
                <v-text-field
                  v-model="password"
                  label="Password"
                  prepend-inner-icon="mdi-lock-outline"
                  type="password"
                  autocomplete="current-password"
                  variant="outlined"
                  density="comfortable"
                  :disabled="loading"
                />

                <v-alert
                  v-if="error"
                  type="error"
                  variant="tonal"
                  density="compact"
                  class="mt-3"
                >
                  {{ error }}
                </v-alert>

                <v-btn
                  type="submit"
                  color="primary"
                  block
                  size="large"
                  class="mt-5"
                  :loading="loading"
                >
                  Sign in
                </v-btn>
              </v-form>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </v-main>
</template>

<script lang="ts" setup>
import { shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const router = useRouter()

const username = shallowRef('')
const password = shallowRef('')
const loading = shallowRef(false)
const error = shallowRef('')

async function onSubmit() {
  if (!username.value || !password.value) return
  error.value = ''
  loading.value = true
  try {
    await authStore.login(username.value, password.value)
    router.push('/listings')
  } catch {
    error.value = 'Invalid username or password.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-main {
  background: rgb(var(--v-theme-background));
}
</style>
