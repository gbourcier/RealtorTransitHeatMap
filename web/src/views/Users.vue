<template>
  <v-container fluid class="pa-2 pa-sm-6">
    <v-card>
      <v-card-title class="d-flex align-center">
        <span>Users</span>
        <v-spacer />
        <v-btn
          color="primary"
          prepend-icon="mdi-account-plus-outline"
          size="small"
          @click="openCreate"
        >
          Add user
        </v-btn>
      </v-card-title>

      <v-divider />

      <v-list>
        <v-list-item
          v-for="u in users"
          :key="u.id"
          :subtitle="u.role"
          :class="{ 'text-disabled': !u.isActive }"
        >
          <template #title>
            <span>{{ u.username }}</span>
            <v-chip
              v-if="!u.isActive"
              size="x-small"
              class="ms-2"
              color="warning"
              variant="tonal"
            >
              inactive
            </v-chip>
            <v-chip
              v-if="u.role === 'admin'"
              size="x-small"
              class="ms-1"
              color="secondary"
              variant="tonal"
            >
              admin
            </v-chip>
          </template>

          <template #append>
            <v-menu location="bottom end" :close-on-content-click="true">
              <template #activator="{ props: menuProps }">
                <v-btn icon="mdi-dots-vertical" variant="text" size="small" v-bind="menuProps" />
              </template>
              <v-list density="compact" min-width="180">
                <v-list-item
                  :title="u.role === 'admin' ? 'Make user' : 'Make admin'"
                  :prepend-icon="u.role === 'admin' ? 'mdi-account-outline' : 'mdi-shield-account-outline'"
                  @click="toggleRole(u)"
                />
                <v-list-item
                  :title="u.isActive ? 'Deactivate' : 'Activate'"
                  :prepend-icon="u.isActive ? 'mdi-account-off-outline' : 'mdi-account-check-outline'"
                  @click="toggleActive(u)"
                />
                <v-list-item
                  title="Reset password"
                  prepend-icon="mdi-lock-reset"
                  @click="openReset(u)"
                />
                <v-divider />
                <v-list-item
                  title="Delete"
                  prepend-icon="mdi-delete-outline"
                  class="text-error"
                  base-color="error"
                  @click="confirmDelete(u)"
                />
              </v-list>
            </v-menu>
          </template>
        </v-list-item>
      </v-list>

      <v-card-text v-if="users.length === 0 && !loadingList" class="text-center text-medium-emphasis py-8">
        No users yet.
      </v-card-text>
    </v-card>
  </v-container>

  <v-dialog v-model="dialog.open" max-width="440" persistent>
    <v-card>
      <v-card-title>{{ dialog.title }}</v-card-title>
      <v-card-text>
        <v-text-field
          v-if="dialog.mode !== 'delete'"
          v-model="dialog.username"
          label="Username"
          variant="outlined"
          density="comfortable"
          :disabled="dialog.mode !== 'create'"
          class="mb-3"
        />
        <v-select
          v-if="dialog.mode === 'create'"
          v-model="dialog.role"
          label="Role"
          :items="roleOptions"
          variant="outlined"
          density="comfortable"
          class="mb-3"
        />
        <v-text-field
          v-if="dialog.mode === 'create' || dialog.mode === 'reset'"
          v-model="dialog.password"
          :label="dialog.mode === 'reset' ? 'New password' : 'Password'"
          type="password"
          variant="outlined"
          density="comfortable"
        />
        <span v-if="dialog.mode === 'delete'">
          Delete <strong>{{ dialog.username }}</strong>? This cannot be undone.
        </span>
        <v-alert v-if="dialog.error" type="error" variant="tonal" density="compact" class="mt-3">
          {{ dialog.error }}
        </v-alert>
      </v-card-text>
      <v-card-actions class="pb-4 px-6">
        <v-spacer />
        <v-btn variant="text" @click="closeDialog">Cancel</v-btn>
        <v-btn
          :color="dialog.mode === 'delete' ? 'error' : 'primary'"
          :loading="dialog.loading"
          @click="submitDialog"
        >
          {{ dialog.mode === 'delete' ? 'Delete' : 'Save' }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts" setup>
import { shallowRef, reactive, onMounted } from 'vue'
import { listUsers, createUser, updateUser, deleteUser } from '../api/users'
import type { User } from '../api/auth'

const users = shallowRef<User[]>([])
const loadingList = shallowRef(false)

const roleOptions = [
  { title: 'User', value: 'user' },
  { title: 'Admin', value: 'admin' },
]

const dialog = reactive({
  open: false,
  mode: 'create' as 'create' | 'reset' | 'delete',
  title: '',
  targetId: '',
  username: '',
  password: '',
  role: 'user' as 'admin' | 'user',
  loading: false,
  error: '',
})

onMounted(async () => {
  loadingList.value = true
  try {
    users.value = await listUsers()
  } finally {
    loadingList.value = false
  }
})

function openCreate() {
  Object.assign(dialog, { open: true, mode: 'create', title: 'Add user', targetId: '', username: '', password: '', role: 'user', error: '' })
}

function openReset(u: User) {
  Object.assign(dialog, { open: true, mode: 'reset', title: 'Reset password', targetId: u.id, username: u.username, password: '', error: '' })
}

function confirmDelete(u: User) {
  Object.assign(dialog, { open: true, mode: 'delete', title: 'Delete user', targetId: u.id, username: u.username, error: '' })
}

function closeDialog() {
  dialog.open = false
}

async function submitDialog() {
  dialog.error = ''
  dialog.loading = true
  try {
    if (dialog.mode === 'create') {
      const u = await createUser({ username: dialog.username, password: dialog.password, role: dialog.role })
      users.value = [...users.value, u]
    } else if (dialog.mode === 'reset') {
      const u = await updateUser(dialog.targetId, { password: dialog.password })
      users.value = users.value.map((x) => (x.id === u.id ? u : x))
    } else {
      await deleteUser(dialog.targetId)
      users.value = users.value.filter((x) => x.id !== dialog.targetId)
    }
    closeDialog()
  } catch (err: any) {
    dialog.error = err?.response?.data?.error ?? 'Operation failed.'
  } finally {
    dialog.loading = false
  }
}

async function toggleRole(u: User) {
  const newRole = u.role === 'admin' ? 'user' : 'admin'
  try {
    const updated = await updateUser(u.id, { role: newRole })
    users.value = users.value.map((x) => (x.id === updated.id ? updated : x))
  } catch (err: any) {
    alert(err?.response?.data?.error ?? 'Failed to update role.')
  }
}

async function toggleActive(u: User) {
  try {
    const updated = await updateUser(u.id, { isActive: !u.isActive })
    users.value = users.value.map((x) => (x.id === updated.id ? updated : x))
  } catch (err: any) {
    alert(err?.response?.data?.error ?? 'Failed to update user.')
  }
}
</script>
