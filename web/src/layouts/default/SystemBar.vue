<template>
  <v-system-bar color="header-bar" height="48" class="header-bar">
    <button
      v-if="onSettingsPage"
      type="button"
      class="back-pill ms-2"
      aria-label="Back to map"
      @click="emit('click:back')"
    >
      <v-icon size="16" class="back-pill__icon">mdi-arrow-left</v-icon>
    </button>

    <v-menu
      v-model="menuOpen"
      location="bottom start"
      offset="18"
      transition="scale-transition"
      :close-on-content-click="true"
    >
      <template #activator="{ props: menuProps }">
        <button
          type="button"
          class="user-pill"
          :class="[onSettingsPage ? 'ms-1' : 'ms-2', { 'user-pill--open': menuOpen }]"
          v-bind="menuProps"
          aria-label="User menu"
        >
          <v-icon size="14" class="user-pill__icon">mdi-account-outline</v-icon>
          <span class="user-pill__label">{{ authStore.user?.username }}</span>
          <v-icon size="12" class="user-pill__chevron">mdi-chevron-down</v-icon>
        </button>
      </template>
      <div class="user-menu" role="menu" aria-label="User menu">
        <header class="user-menu__header">
          <v-icon size="22" class="user-menu__avatar">mdi-account-circle-outline</v-icon>
          <div class="user-menu__identity">
            <span class="user-menu__name">{{ authStore.user?.username }}</span>
            <span class="user-menu__role">{{ authStore.isAdmin ? 'Admin' : 'User' }}</span>
          </div>
        </header>
        <div class="user-menu__items">
          <button
            v-if="authStore.isAdmin"
            type="button"
            role="menuitem"
            class="user-menu__item"
            :class="{ 'user-menu__item--active': settingsActive }"
            @click="emit('click:settings')"
          >
            <v-icon size="18" class="user-menu__item-icon">mdi-cog-outline</v-icon>
            <span class="user-menu__item-label">Settings</span>
          </button>
          <button
            type="button"
            role="menuitem"
            class="user-menu__item"
            @click="onLogout"
          >
            <v-icon size="18" class="user-menu__item-icon">mdi-logout</v-icon>
            <span class="user-menu__item-label">Logout</span>
          </button>
        </div>
      </div>
    </v-menu>

    <v-spacer />

    <div id="header-filters-slot" class="header-bar__filters" />
    <div id="header-actions-slot" class="header-bar__actions" />
  </v-system-bar>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

withDefaults(defineProps<{ settingsActive?: boolean; onSettingsPage?: boolean }>(), {
  settingsActive: false,
  onSettingsPage: false,
})
const emit = defineEmits(['click:settings', 'click:back'])

const authStore = useAuthStore()
const router = useRouter()
const menuOpen = ref(false)

async function onLogout() {
  await authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.header-bar :deep(.v-system-bar__content),
.header-bar.v-system-bar {
  padding-inline: 0;
}

.header-bar__filters {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  flex: 0 0 auto;
}

.header-bar__actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: 0 0 auto;
}

.header-bar__actions {
  margin-right: 8px;
}

.header-bar__actions:not(:empty) {
  margin-left: 6px;
}

.back-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 28px;
  width: 28px;
  padding: 0;
  border-radius: 999px;
  border: 1px solid rgba(var(--v-theme-primary), 0.7);
  background: transparent;
  color: rgb(var(--v-theme-primary));
  cursor: pointer;
  transition: background-color 120ms ease, border-color 120ms ease, color 120ms ease;
}

.back-pill:hover {
  background-color: rgba(var(--v-theme-primary), 0.08);
  border-color: rgb(var(--v-theme-primary));
}

.back-pill:focus-visible {
  outline: 2px solid rgb(var(--v-theme-primary));
  outline-offset: 2px;
}

.back-pill__icon {
  opacity: 0.95;
}

.user-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  border: 1px solid rgba(var(--v-theme-on-surface), 0.22);
  background: transparent;
  color: rgba(var(--v-theme-on-surface), 0.88);
  font-size: 0.75rem;
  font-weight: 500;
  letter-spacing: normal;
  cursor: pointer;
  transition: background-color 120ms ease, border-color 120ms ease;
}

.user-pill:hover {
  background-color: rgba(var(--v-theme-on-surface), 0.05);
  border-color: rgba(var(--v-theme-on-surface), 0.32);
}

.user-pill--open {
  background-color: rgba(var(--v-theme-on-surface), 0.12);
  border-color: rgba(var(--v-theme-on-surface), 0.4);
}

.user-pill:focus-visible {
  outline: 2px solid rgb(var(--v-theme-primary));
  outline-offset: 2px;
}

.user-pill__icon {
  opacity: 0.9;
}

.user-pill__chevron {
  opacity: 0.6;
}

.user-menu {
  width: 220px;
  border-radius: 16px;
  background-color: rgb(var(--v-theme-surface));
  border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
  box-shadow: 0 16px 40px rgba(var(--v-theme-shadow), 0.5);
  color: rgba(var(--v-theme-on-surface), 0.92);
  overflow: hidden;
}

.user-menu__header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
}

.user-menu__avatar {
  color: rgba(var(--v-theme-on-surface), 0.7);
}

.user-menu__identity {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.user-menu__name {
  font-size: 0.9375rem;
  font-weight: 600;
  line-height: 1.2;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-menu__role {
  font-size: 0.75rem;
  color: rgba(var(--v-theme-on-surface), 0.6);
}

.user-menu__items {
  padding: 6px;
  border-top: 1px solid rgba(var(--v-theme-on-surface), 0.06);
}

.user-menu__item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 9px 10px;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: rgba(var(--v-theme-on-surface), 0.88);
  font-size: 0.875rem;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  transition: background-color 120ms ease, color 120ms ease;
}

.user-menu__item:hover {
  background-color: rgba(var(--v-theme-on-surface), 0.06);
}

.user-menu__item:focus-visible {
  outline: 2px solid rgb(var(--v-theme-primary));
  outline-offset: -2px;
}

.user-menu__item--active {
  color: rgb(var(--v-theme-primary));
}

.user-menu__item--active:hover {
  background-color: rgba(var(--v-theme-primary), 0.08);
}

.user-menu__item-icon {
  opacity: 0.9;
}
</style>
