<template>
  <v-system-bar
    color="header-bar"
    height="66"
    class="header-bar"
    :class="{ 'header-bar--with-panel': showPanelToggle }"
  >
    <button
      v-if="showBack"
      type="button"
      class="back-pill"
      aria-label="Back to map"
      @click="emit('click:back')"
    >
      <v-icon size="18" class="back-pill__icon">mdi-arrow-left</v-icon>
    </button>

    <RouterLink to="/listings" class="brand" aria-label="HouseMap" @click="emit('click:brand')">
      <span class="brand__tile">
        <img src="/brand/sona-logo.svg" alt="" class="brand__logo" />
      </span>
      <span class="brand__word">House<span>Map</span></span>
    </RouterLink>

    <v-spacer />

    <div id="header-filters-slot" class="header-bar__filters" />
    <div class="header-bar__divider" />

    <v-menu
      v-model="menuOpen"
      :location="mobile ? 'bottom center' : 'bottom end'"
      offset="10"
      :transition="mobile ? 'mobile-sheet-transition' : 'scale-transition'"
      :content-class="mobile ? 'mobile-sheet-menu mobile-sheet-menu--user' : undefined"
      :close-on-content-click="false"
    >
      <template #activator="{ props: menuProps }">
        <button
          type="button"
          class="user-pill"
          :class="{ 'user-pill--open': menuOpen }"
          v-bind="menuProps"
          aria-label="User menu"
        >
          <v-icon size="18" class="user-pill__icon">mdi-account-outline</v-icon>
          <span class="user-pill__label">{{ authStore.user?.username }}</span>
          <v-icon size="15" class="user-pill__chevron">mdi-chevron-down</v-icon>
        </button>
      </template>

      <div
        v-if="mobile && menuView === 'settings'"
        class="user-menu"
        :class="mobile ? menuDragClasses : undefined"
        :style="mobile ? menuDragStyle : undefined"
        role="menu"
        aria-label="Settings"
      >
        <button
          v-if="mobile"
          type="button"
          class="mobile-sheet-grip"
          aria-label="Drag down to close settings"
          @pointerdown="onMenuDragPointerDown"
          @pointermove="onMenuDragPointerMove"
          @pointerup="onMenuDragPointerUp"
          @pointercancel="onMenuDragPointerCancel"
        />
        <header class="user-menu__header">
          <button
            type="button"
            class="user-menu__back"
            aria-label="Back to user menu"
            @click="menuView = 'user'"
          >
            <v-icon size="20">mdi-arrow-left</v-icon>
          </button>
          <div class="user-menu__identity">
            <span class="user-menu__name">Settings</span>
            <span class="user-menu__role">Administration</span>
          </div>
        </header>

        <div class="user-menu__sep" />

        <div class="user-menu__items">
          <button
            v-for="item in settingsItems"
            :key="item.to"
            type="button"
            role="menuitem"
            class="user-menu__item user-menu__item--setting"
            :class="{ 'user-menu__item--active': route.path === item.to }"
            @click="onSettingRoute(item.to)"
          >
            <v-icon size="20" class="user-menu__item-icon">{{ item.icon }}</v-icon>
            <span class="user-menu__setting-copy">
              <span class="user-menu__item-label">{{ item.title }}</span>
              <span class="user-menu__setting-subtitle">{{ item.subtitle }}</span>
            </span>
          </button>
        </div>
      </div>

      <div
        v-else
        class="user-menu"
        :class="mobile ? menuDragClasses : undefined"
        :style="mobile ? menuDragStyle : undefined"
        role="menu"
        aria-label="User menu"
      >
        <button
          v-if="mobile"
          type="button"
          class="mobile-sheet-grip"
          aria-label="Drag down to close user menu"
          @pointerdown="onMenuDragPointerDown"
          @pointermove="onMenuDragPointerMove"
          @pointerup="onMenuDragPointerUp"
          @pointercancel="onMenuDragPointerCancel"
        />
        <header class="user-menu__header">
          <span class="user-menu__avatar">
            <v-icon size="22">mdi-account-outline</v-icon>
          </span>
          <div class="user-menu__identity">
            <span class="user-menu__name">{{ authStore.user?.username }}</span>
            <span class="user-menu__role">{{ authStore.isAdmin ? 'Admin' : 'User' }}</span>
          </div>
        </header>

        <div class="user-menu__sep" />

        <div class="user-menu__items">
          <button
            type="button"
            role="menuitem"
            class="user-menu__item"
            :class="{ 'user-menu__item--active': favoritesActive }"
            @click="onFavorites"
          >
            <v-icon size="18" class="user-menu__item-icon">mdi-heart-outline</v-icon>
            <span class="user-menu__item-label">Manage Favorites</span>
          </button>
          <button
            v-if="authStore.isAdmin"
            type="button"
            role="menuitem"
            class="user-menu__item"
            :class="{ 'user-menu__item--active': settingsActive }"
            @click="onSettings"
          >
            <v-icon size="18" class="user-menu__item-icon">mdi-cog-outline</v-icon>
            <span class="user-menu__item-label">Settings</span>
          </button>
          <button
            v-if="showPanelToggle"
            type="button"
            role="menuitem"
            class="user-menu__item user-menu__item--mobile-only"
            @click="onPanelToggle"
          >
            <v-icon size="18" class="user-menu__item-icon">mdi-dock-right</v-icon>
            <span class="user-menu__item-label">Toggle results panel</span>
          </button>
        </div>

        <div class="user-menu__sep" />

        <div class="user-menu__items">
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

    <div id="header-actions-slot" class="header-bar__actions" />
  </v-system-bar>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue'
import { useDisplay } from 'vuetify'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useBottomSheetDrag } from '../../composables/useBottomSheetDrag'
import { useMobileBottomSheetCoordinator } from '../../composables/useMobileBottomSheetCoordinator'
import { useAuthStore } from '../../stores/auth'
import { useSavedFiltersStore } from '../../stores/savedFilters'

withDefaults(defineProps<{ settingsActive?: boolean; showBack?: boolean; showPanelToggle?: boolean }>(), {
  settingsActive: false,
  showBack: false,
  showPanelToggle: false,
})
const emit = defineEmits(['click:settings', 'click:back', 'click:panel-toggle', 'click:brand'])

const authStore = useAuthStore()
const savedFiltersStore = useSavedFiltersStore()
const router = useRouter()
const route = useRoute()
const { mobile } = useDisplay()
const menuOpen = ref(false)
const menuView = ref<'user' | 'settings'>('user')
const {
  dragClasses: menuDragClasses,
  dragStyle: menuDragStyle,
  resetDrag: resetMenuDrag,
  onDragPointerDown: onMenuDragPointerDown,
  onDragPointerMove: onMenuDragPointerMove,
  onDragPointerUp: onMenuDragPointerUp,
  onDragPointerCancel: onMenuDragPointerCancel,
} = useBottomSheetDrag(() => {
  menuOpen.value = false
})
useMobileBottomSheetCoordinator({
  open: menuOpen,
  close: () => {
    menuOpen.value = false
  },
  enabled: () => mobile.value,
})

const settingsItems = [
  {
    title: 'Realtor Scraper',
    subtitle: 'Schedules & manual runs',
    icon: 'mdi-cog-transfer-outline',
    to: '/scraper',
  },
  {
    title: 'GTFS Data',
    subtitle: 'Transit feed refresh',
    icon: 'mdi-train',
    to: '/transit',
  },
  {
    title: 'Users',
    subtitle: 'Manage accounts',
    icon: 'mdi-account-group-outline',
    to: '/users',
  },
]

const favoritesActive = computed(() => route.name === 'favorites')

watch(menuOpen, (open) => {
  if (open) resetMenuDrag()
  if (!open) {
    menuView.value = 'user'
  }
})

function onFavorites() {
  menuOpen.value = false
  router.push('/favorites')
}

function onSettings() {
  if (mobile.value) {
    menuView.value = 'settings'
    return
  }
  menuOpen.value = false
  emit('click:settings')
}

function onSettingRoute(path: string) {
  menuOpen.value = false
  router.push(path)
}

function onPanelToggle() {
  menuOpen.value = false
  emit('click:panel-toggle')
}

async function onLogout() {
  menuOpen.value = false
  await authStore.logout()
  savedFiltersStore.reset()
  router.push('/login')
}
</script>

<style scoped>
.header-bar :deep(.v-system-bar__content),
.header-bar.v-system-bar {
  padding-inline: 0;
  overflow: visible;
}

.header-bar {
  position: relative;
  z-index: 1000;
  border-bottom: 1px solid rgba(0, 0, 0, 0.5);
  box-shadow: 0 8px 26px -14px #000;
  font-family: Inter, system-ui, sans-serif;
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 11px;
  flex: 0 0 auto;
  margin-left: 20px;
  color: #f4f1e8;
  text-decoration: none;
}

.back-pill + .brand {
  margin-left: 10px;
}

.brand__tile {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  border-radius: 12px;
  background: linear-gradient(160deg, #fff, #eaf2fb);
  box-shadow: inset 0 0 0 1px rgba(21, 35, 61, 0.1), 0 2px 6px rgba(0, 0, 0, 0.3);
}

.brand__logo {
  display: block;
  width: 30px;
  height: 30px;
}

.brand__word {
  font-family: "Baloo 2", Inter, system-ui, sans-serif;
  font-size: 21px;
  font-weight: 800;
  letter-spacing: -0.01em;
  line-height: 1;
}

.brand__word span {
  color: #b6f24a;
}

.header-bar__filters {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
  flex: 0 0 auto;
}

.header-bar__divider {
  width: 1px;
  height: 26px;
  margin-inline: 12px;
  background: rgba(244, 241, 232, 0.12);
  flex: 0 0 auto;
}

.header-bar__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-right: 8px;
  flex: 0 0 auto;
}

.back-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 40px;
  width: 40px;
  margin-left: 20px;
  padding: 0;
  border-radius: 11px;
  border: 1px solid rgba(244, 241, 232, 0.12);
  background: #2a2d27;
  color: #f4f1e8;
  cursor: pointer;
  transition: background-color 140ms ease, border-color 140ms ease, transform 60ms ease;
}

.back-pill:hover {
  background-color: #34382f;
  border-color: rgba(244, 241, 232, 0.22);
}

.back-pill:active {
  transform: translateY(1px);
}

.back-pill:focus-visible {
  outline: 2px solid #6ccff6;
  outline-offset: 2px;
}

.user-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 40px;
  margin-right: 20px;
  padding: 0 14px;
  border-radius: 999px;
  border: 1px solid rgba(244, 241, 232, 0.12);
  background: #2a2d27;
  color: #f4f1e8;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0;
  cursor: pointer;
  transition: background-color 140ms ease, border-color 140ms ease, transform 60ms ease;
}

.header-bar--with-panel .user-pill {
  margin-right: 8px;
}

.user-pill:hover,
.user-pill--open {
  background-color: #34382f;
  border-color: rgba(244, 241, 232, 0.22);
}

.user-pill:active {
  transform: translateY(1px);
}

.user-pill:focus-visible {
  outline: 2px solid #6ccff6;
  outline-offset: 2px;
}

.user-pill__chevron {
  color: rgba(244, 241, 232, 0.52);
}

.user-menu {
  position: relative;
  width: 260px;
  padding: 8px;
  border-radius: 16px;
  background-color: #262925;
  border: 1px solid rgba(244, 241, 232, 0.12);
  box-shadow: 0 24px 60px -18px rgba(0, 0, 0, 0.8);
  color: #f4f1e8;
  overflow: hidden;
  font-family: Inter, system-ui, sans-serif;
}

.user-menu__header {
  display: flex;
  align-items: center;
  gap: 13px;
  padding: 12px 12px 10px;
}

.user-menu__avatar {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  border-radius: 999px;
  background: #2a2d27;
  border: 1px solid rgba(244, 241, 232, 0.12);
  color: #f4f1e8;
}

.user-menu__back {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  padding: 0;
  border-radius: 999px;
  border: 1px solid rgba(244, 241, 232, 0.12);
  background: #2a2d27;
  color: #f4f1e8;
  cursor: pointer;
}

.user-menu__back:focus-visible {
  outline: 2px solid #6ccff6;
  outline-offset: 2px;
}

.user-menu__identity {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.user-menu__name {
  font-size: 16px;
  font-weight: 700;
  line-height: 1.2;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-menu__role {
  margin-top: 1px;
  font-size: 13px;
  color: rgba(244, 241, 232, 0.52);
}

.user-menu__sep {
  height: 1px;
  margin: 7px 6px;
  background: rgba(244, 241, 232, 0.12);
}

.user-menu__items {
  display: flex;
  flex-direction: column;
}

.user-menu__item {
  display: flex;
  align-items: center;
  gap: 13px;
  width: 100%;
  height: 46px;
  padding: 0 12px;
  border: 0;
  border-radius: 11px;
  background: transparent;
  color: #f4f1e8;
  font-size: 15px;
  font-weight: 600;
  text-align: left;
  cursor: pointer;
  transition: background-color 120ms ease, color 120ms ease;
}

.user-menu__item:hover {
  background-color: rgba(244, 241, 232, 0.06);
}

.user-menu__item:focus-visible {
  outline: 2px solid #6ccff6;
  outline-offset: -2px;
}

.user-menu__item--active {
  color: #b6f24a;
  background: rgba(182, 242, 74, 0.12);
}

.user-menu__item-icon {
  color: rgba(244, 241, 232, 0.52);
}

.user-menu__item--setting {
  height: 62px;
}

.user-menu__setting-copy {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.user-menu__setting-subtitle {
  margin-top: 2px;
  color: rgba(244, 241, 232, 0.52);
  font-size: 13px;
  font-weight: 400;
  line-height: 1.2;
}

.user-menu__item--mobile-only {
  display: none;
}

@media (max-width: 899px) {
  :global(.mobile-sheet-menu.v-overlay__content) {
    position: fixed !important;
    top: auto !important;
    right: 11px !important;
    bottom: 0 !important;
    left: 11px !important;
    width: auto !important;
    min-width: 0 !important;
    max-width: none !important;
  }

  :global(.mobile-sheet-transition-enter-active),
  :global(.mobile-sheet-transition-leave-active) {
    transition: opacity 180ms ease, transform 260ms cubic-bezier(0.22, 0.7, 0.3, 1);
  }

  :global(.mobile-sheet-transition-enter-from),
  :global(.mobile-sheet-transition-leave-to) {
    opacity: 0;
    transform: translateY(100%);
  }

  .header-bar.v-system-bar {
    height: 58px !important;
  }

  .brand {
    margin-left: 11px;
  }

  .back-pill {
    width: 38px;
    height: 38px;
    margin-left: 11px;
  }

  .back-pill + .brand {
    margin-left: 8px;
  }

  .brand__tile {
    width: 40px;
    height: 40px;
  }

  .brand__logo {
    width: 31px;
    height: 31px;
  }

  .brand__word,
  .header-bar__actions {
    display: none;
  }

  .header-bar__filters {
    gap: 11px;
  }

  .header-bar__divider {
    display: block;
    height: 28px;
    margin-inline: 10px;
  }

  .user-pill {
    width: 38px;
    height: 38px;
    margin-left: 0;
    margin-right: 11px;
    padding: 0;
    justify-content: center;
  }

  .user-pill__label,
  .user-pill__chevron {
    display: none;
  }

  .user-menu {
    width: 100%;
    max-height: 88dvh;
    overflow-y: auto;
    border-radius: 22px 22px 0 0;
  }

  .user-menu__item--mobile-only {
    display: flex;
  }
}
</style>
