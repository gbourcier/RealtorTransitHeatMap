<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, onActivated, onDeactivated, nextTick, provide, watch } from "vue";

defineOptions({ name: "Listings" });
import { useDisplay } from "vuetify";
import type { Listing } from "../api/listings";
import ListingsMap from "../components/ListingsMap.vue";
import ListingFilters from "../components/ListingFilters.vue";
import SaveFilterDialog from "../components/SaveFilterDialog.vue";
import ViewSelector from "../components/ViewSelector.vue";
import FiltersButton from "../components/FiltersButton.vue";
import ListingsCountPill from "../components/ListingsCountPill.vue";
import ListingsSidePanel from "../components/ListingsSidePanel.vue";
import ListingsMobileList from "../components/ListingsMobileList.vue";
import MapLegend from "../components/MapLegend.vue";
import MobileViewToggle from "../components/MobileViewToggle.vue";
import { useListingFilters } from "../composables/useListingFilters";
import { useSavedViews } from "../composables/useSavedViews";
import { useListings } from "../composables/useListings";
import { useFavorites, favoritesKey } from "../composables/useFavorites";
import { useBodyScrollLock } from "../composables/useBodyScrollLock";
import { useSavedFiltersStore } from "../stores/savedFilters";
import { debounce } from "../utils/debounce";

const DEFAULT_PANEL_WIDTH = 360;
const PANEL_WIDTH_STORAGE_KEY = "listingsSidePanelWidth";

const { mdAndUp } = useDisplay();
const viewMode = ref<"list" | "map">("map");
const drawerOpen = ref(true);
const panelWidth = ref(readStoredPanelWidth());
const filtersOpen = ref(false);
const filtersPanelRef = ref<{ rootEl: HTMLElement | null } | null>(null);
const saveModalOpen = ref(false);
const teleportReady = ref(false);
const headerVisible = ref(true);
let headerTeleportFrame: number | null = null;

function cancelHeaderTeleportSync(): void {
    if (headerTeleportFrame === null) return;
    window.cancelAnimationFrame(headerTeleportFrame);
    headerTeleportFrame = null;
}

function syncHeaderTeleportTargets(): void {
    cancelHeaderTeleportSync();
    const ready =
        !!document.getElementById("header-filters-slot") &&
        !!document.getElementById("header-actions-slot");

    teleportReady.value = ready;
    if (!ready && headerVisible.value) {
        headerTeleportFrame = window.requestAnimationFrame(syncHeaderTeleportTargets);
    }
}

async function refreshHeaderTeleportTargets(): Promise<void> {
    await nextTick();
    syncHeaderTeleportTargets();
}

onActivated(() => {
    headerVisible.value = true;
    void refreshHeaderTeleportTargets();
});
onDeactivated(() => {
    headerVisible.value = false;
    teleportReady.value = false;
    cancelHeaderTeleportSync();
});

const filters = useListingFilters();
const savedFiltersStore = useSavedFiltersStore();
const views = useSavedViews(filters);
if (savedFiltersStore.defaultFilter) {
    views.applySaved(savedFiltersStore.defaultFilter);
}
function openSaveFromHeader(): void {
    saveModalOpen.value = true;
}
const listings = useListings(filters);
const favorites = useFavorites();
provide(favoritesKey, favorites);

function onMapToggleFavorite(payload: {
    board: number;
    mls: number;
    isFavorite: boolean;
}): void {
    favorites.toggle(payload);
    const next = favorites.isFavorite(
        payload.board,
        payload.mls,
        payload.isFavorite,
    );
    mapRef.value?.setFavorite(payload.board, payload.mls, next);
}

function flushFavorites(): void {
    favorites.flush();
}

function onVisibilityChange(): void {
    if (document.visibilityState === "hidden") flushFavorites();
}

function onDocumentPointerDown(event: PointerEvent): void {
    if (!filtersOpen.value || !headerVisible.value) return;
    const target = event.target;
    if (!(target instanceof Element)) return;
    if (target.closest(".filters-btn")) return;

    const panelEl = filtersPanelRef.value?.rootEl;
    if (panelEl?.contains(target)) return;

    filtersOpen.value = false;
}

function toggleResultsPanel(): void {
    drawerOpen.value = !drawerOpen.value;
}

function selectToolbarBuildingType(id: number | null): void {
    if (id == null) {
        filters.buildingTypes.value = [];
        return;
    }
    filters.buildingTypes.value =
        filters.buildingTypes.value.length === 1 && filters.buildingTypes.value[0] === id
            ? []
            : [id];
}

const mapRef = ref<InstanceType<typeof ListingsMap> | null>(null);
const mapCount = ref(0);
const mapLoading = ref(false);
const selectedKey = ref<string | null>(null);

const errorToast = ref<string | null>(null);
const errorToastOpen = ref(false);

function showError(message: string | null): void {
    if (!message) return;
    errorToast.value = message;
    errorToastOpen.value = true;
}

watch(() => listings.error.value, showError);
favorites.onError(showError);

useBodyScrollLock();

const reloadListings = debounce(() => listings.loadInitial(), 250);

watch(
    () => [
        filters.buildingTypes.value.join(","),
        filters.maxPrice.value,
        filters.maxCommuteSec.value,
        filters.newWithinDays.value,
        filters.minBedrooms.value,
        filters.minBathrooms.value,
        filters.minInteriorAreaSqft.value,
        filters.favoritesOnly.value,
        filters.includeExpired.value,
    ],
    reloadListings,
);

function listingKey(item: Listing): string {
    return `${item.board}-${item.mls}`;
}

function focusListingOnMap(item: Listing): void {
    selectedKey.value = listingKey(item);
    mapRef.value?.focusListing(item.board, item.mls);
}

function highlightListingOnMap(item: Listing): void {
    mapRef.value?.highlightListing(item.board, item.mls);
}

function clearMapHighlight(): void {
    mapRef.value?.clearHighlight();
}

function openListing(item: Listing): void {
    if (!item.slug) return;
    window.open(item.slug, "_blank", "noopener,noreferrer");
}

function readStoredPanelWidth(): number {
    if (typeof window === "undefined") return DEFAULT_PANEL_WIDTH;
    const stored = Number(window.localStorage.getItem(PANEL_WIDTH_STORAGE_KEY));
    return Number.isFinite(stored) && stored > 0 ? stored : DEFAULT_PANEL_WIDTH;
}

function updatePanelWidth(width: number): void {
    panelWidth.value = width;
    window.localStorage.setItem(PANEL_WIDTH_STORAGE_KEY, String(width));
}

onMounted(() => {
    void refreshHeaderTeleportTargets();
    listings.loadInitial();
    window.addEventListener("beforeunload", flushFavorites);
    window.addEventListener("listings:toggle-panel", toggleResultsPanel);
    document.addEventListener("pointerdown", onDocumentPointerDown);
    document.addEventListener("visibilitychange", onVisibilityChange);
});

onBeforeUnmount(() => {
    cancelHeaderTeleportSync();
    reloadListings.cancel();
    window.removeEventListener("beforeunload", flushFavorites);
    window.removeEventListener("listings:toggle-panel", toggleResultsPanel);
    document.removeEventListener("pointerdown", onDocumentPointerDown);
    document.removeEventListener("visibilitychange", onVisibilityChange);
    flushFavorites();
});
</script>

<template>
    <Teleport v-if="teleportReady" to="#header-filters-slot">
        <template v-if="headerVisible">
            <ListingsCountPill :total="listings.total.value" :loading="listings.loading.value || mapLoading" />
            <ViewSelector :views="views" :filters="filters" @save="openSaveFromHeader" />
            <FiltersButton
                :open="filtersOpen"
                :count="filters.activeFilterCount.value"
                @toggle="filtersOpen = !filtersOpen"
            />
        </template>
    </Teleport>

    <Transition name="filter-drawer">
        <ListingFilters
            ref="filtersPanelRef"
            v-if="filtersOpen && headerVisible"
            :state="filters"
            :views="views"
            :total="listings.total.value"
            @close="filtersOpen = false"
            @error="showError"
            @save="saveModalOpen = true"
        />
    </Transition>

    <SaveFilterDialog
        v-if="headerVisible"
        v-model="saveModalOpen"
        :filters="filters"
        :views="views"
    />

    <Teleport v-if="teleportReady" to="#header-actions-slot">
        <button
            v-if="headerVisible && mdAndUp"
            type="button"
            class="panel-toggle"
            :class="{ 'panel-toggle--on': drawerOpen }"
            :aria-pressed="drawerOpen"
            :aria-label="drawerOpen ? 'Hide results panel' : 'Show results panel'"
            @click="drawerOpen = !drawerOpen"
        >
            <v-icon size="18">mdi-dock-right</v-icon>
        </button>
    </Teleport>

    <div
        v-show="mdAndUp || viewMode === 'map'"
        class="map-fullbleed"
        :class="{ 'map-fullbleed--with-panel': mdAndUp && drawerOpen }"
        :style="{ '--listings-panel-width': `${panelWidth}px` }"
    >
        <ListingsMap
            ref="mapRef"
            class="map-fullbleed__map"
            :class="{ 'map-fullbleed__map--dim': mapLoading }"
            :max-price="filters.maxPrice.value"
            :building-types="filters.buildingTypes.value"
            :max-commute-sec="filters.maxCommuteSec.value"
            :new-within-days="filters.newWithinDays.value"
            :min-bedrooms="filters.minBedrooms.value"
            :min-bathrooms="filters.minBathrooms.value"
            :min-interior-area-sqft="filters.minInteriorAreaSqft.value"
            :favorites-only="filters.favoritesOnly.value"
            :include-expired="filters.includeExpired.value"
            @update:count="mapCount = $event"
            @update:loading="mapLoading = $event"
            @toggle-favorite="onMapToggleFavorite"
            @error="showError"
        />
        <div
            v-if="mapLoading"
            class="map-fullbleed__loading"
            aria-live="polite"
            aria-label="Loading map listings"
        >
            <v-progress-circular indeterminate size="18" width="3" color="primary" />
            <span>Loading listings…</span>
        </div>
        <ListingsSidePanel
            v-if="mdAndUp && drawerOpen"
            :items="listings.items.value"
            :loading="listings.loading.value"
            :has-more="listings.hasMore.value"
            :sort-by="listings.sortBy.value"
            :sort-dir="listings.sortDir.value"
            :sort-options="listings.sortOptions"
            :building-types="filters.buildingTypes.value"
            :selected-key="selectedKey"
            :width="panelWidth"
            @select-sort="listings.selectSort"
            @select-building-type="selectToolbarBuildingType"
            @update:width="updatePanelWidth"
            @card-click="focusListingOnMap"
            @card-hover="highlightListingOnMap"
            @card-leave="clearMapHighlight"
            @load-more="listings.loadMore"
        />
    </div>

    <ListingsMobileList
        v-if="!mdAndUp && viewMode === 'list'"
        :items="listings.items.value"
        :loading="listings.loading.value"
        :has-more="listings.hasMore.value"
        :sort-by="listings.sortBy.value"
        :sort-dir="listings.sortDir.value"
        :sort-options="listings.sortOptions"
        :building-types="filters.buildingTypes.value"
        @select-sort="listings.selectSort"
        @select-building-type="selectToolbarBuildingType"
        @card-click="openListing"
        @load-more="listings.loadMore"
    />

    <MapLegend v-if="!mdAndUp && viewMode === 'map'" />

    <MobileViewToggle v-if="!mdAndUp" v-model="viewMode" />

    <v-snackbar
        v-model="errorToastOpen"
        color="error"
        location="top"
        :timeout="6000"
    >
        {{ errorToast }}
        <template #actions>
            <v-btn variant="text" @click="errorToastOpen = false">Dismiss</v-btn>
        </template>
    </v-snackbar>

    <v-snackbar
        :model-value="favorites.snackbar.value.open"
        location="bottom"
        :timeout="-1"
        class="favorites-snackbar"
        content-class="favorites-snackbar__content"
    >
        {{ favorites.snackbar.value.count === 1
            ? "Removed from favorites"
            : `Removed ${favorites.snackbar.value.count} from favorites` }}
        <template #actions>
            <v-btn variant="text" color="primary" @click="favorites.undo()">Undo</v-btn>
        </template>
    </v-snackbar>
</template>

<style scoped>
.map-fullbleed {
    position: relative;
    height: calc(100dvh - 66px);
    width: 100%;
    display: flex;
}

.map-fullbleed__map {
    flex: 1 1 auto;
    min-height: 0;
    min-width: 0;
    transition: filter 300ms ease;
}

.map-fullbleed__map--dim {
    filter: brightness(0.78);
}

.map-fullbleed__loading {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 11px;
    pointer-events: none;
    z-index: 500;
    color: #f4f1e8;
    font-size: 14px;
    font-weight: 600;
}

.map-fullbleed__loading {
    top: 54%;
    bottom: auto;
    left: 50%;
    right: auto;
    padding: 11px 20px;
    border-radius: 999px;
    background: rgba(24, 26, 23, 0.9);
    border: 1px solid rgba(244, 241, 232, 0.12);
    box-shadow: 0 14px 40px -16px #000;
    transform: translate(-50%, -50%);
}

.map-fullbleed--with-panel .map-fullbleed__loading {
    margin-left: calc(var(--listings-panel-width) / -2);
}

.filter-drawer-enter-active,
.filter-drawer-leave-active {
    transition: opacity 160ms ease, transform 160ms ease;
}

.filter-drawer-enter-from,
.filter-drawer-leave-to {
    opacity: 0;
    transform: translateX(12px);
}

.panel-toggle {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    padding: 0;
    border-radius: 999px;
    border: 1px solid transparent;
    background: transparent;
    color: #f4f1e8;
    cursor: pointer;
    transition: background-color 140ms ease, border-color 140ms ease, transform 60ms ease;
}

.panel-toggle:hover {
    background: rgba(244, 241, 232, 0.06);
}

.panel-toggle--on {
    background: #2a2d27;
    border-color: rgba(244, 241, 232, 0.12);
}

.panel-toggle--on:hover {
    background: #34382f;
    border-color: rgba(244, 241, 232, 0.22);
}

.panel-toggle:active {
    transform: translateY(1px);
}

.panel-toggle:focus-visible {
    outline: 2px solid #6ccff6;
    outline-offset: 2px;
}

.favorites-snackbar :deep(.favorites-snackbar__content) {
    background: rgb(var(--v-theme-popup-overlay));
    color: rgb(var(--v-theme-on-surface));
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-radius: 10px;
    box-shadow:
        0 18px 48px -18px rgba(var(--v-theme-shadow), 0.85),
        inset 0 1px 0 rgba(var(--v-theme-on-surface), 0.04);
}

.favorites-snackbar :deep(.v-snackbar__actions) {
    color: rgb(var(--v-theme-primary));
}

@media (max-width: 899px) {
    .map-fullbleed {
        height: calc(100dvh - 58px);
    }
}
</style>
