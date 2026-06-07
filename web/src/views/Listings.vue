<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, onActivated, onDeactivated, provide, watch } from "vue";

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

const { mdAndUp } = useDisplay();
const viewMode = ref<"list" | "map">("map");
const drawerOpen = ref(true);
const filtersOpen = ref(false);
const saveModalOpen = ref(false);
const teleportReady = ref(false);
const headerVisible = ref(true);

onActivated(() => (headerVisible.value = true));
onDeactivated(() => (headerVisible.value = false));

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

onMounted(() => {
    teleportReady.value = true;
    listings.loadInitial();
    window.addEventListener("beforeunload", flushFavorites);
    document.addEventListener("visibilitychange", onVisibilityChange);
});

onBeforeUnmount(() => {
    reloadListings.cancel();
    window.removeEventListener("beforeunload", flushFavorites);
    document.removeEventListener("visibilitychange", onVisibilityChange);
    flushFavorites();
});
</script>

<template>
    <Teleport to="#header-filters-slot" :disabled="!teleportReady">
        <template v-if="headerVisible">
            <ListingsCountPill :total="listings.total.value" />
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

    <Teleport to="#header-actions-slot" :disabled="!teleportReady">
        <v-btn
            v-if="headerVisible && mdAndUp"
            icon
            variant="text"
            size="x-small"
            :active="drawerOpen"
            :aria-label="drawerOpen ? 'Hide results panel' : 'Show results panel'"
            @click="drawerOpen = !drawerOpen"
        >
            <v-icon size="20">mdi-dock-right</v-icon>
        </v-btn>
    </Teleport>

    <div
        v-show="mdAndUp || viewMode === 'map'"
        class="map-fullbleed"
        :class="{ 'map-fullbleed--with-panel': mdAndUp && drawerOpen }"
    >
        <ListingsMap
            ref="mapRef"
            class="map-fullbleed__map"
            :max-price="filters.maxPrice.value"
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
            <v-progress-circular indeterminate size="72" width="5" color="primary" />
        </div>
        <ListingsSidePanel
            v-if="mdAndUp && drawerOpen"
            :items="listings.items.value"
            :loading="listings.loading.value"
            :has-more="listings.hasMore.value"
            :sort-by="listings.sortBy.value"
            :sort-dir="listings.sortDir.value"
            :sort-options="listings.sortOptions"
            :selected-key="selectedKey"
            @select-sort="listings.selectSort"
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
        @select-sort="listings.selectSort"
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
    >
        {{ favorites.snackbar.value.count === 1
            ? "Removed from favorites"
            : `Removed ${favorites.snackbar.value.count} from favorites` }}
        <template #actions>
            <v-btn variant="text" color="accent" @click="favorites.undo()">Undo</v-btn>
        </template>
    </v-snackbar>
</template>

<style scoped>
.map-fullbleed {
    position: relative;
    height: calc(100dvh - 40px);
    width: 100%;
    display: flex;
}

.map-fullbleed__map {
    flex: 1 1 auto;
    min-height: 0;
    min-width: 0;
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
    pointer-events: none;
    z-index: 500;
}

.map-fullbleed--with-panel .map-fullbleed__loading {
    right: 360px;
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
</style>
