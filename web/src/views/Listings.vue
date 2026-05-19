<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { useDisplay } from "vuetify";
import type { Listing } from "../api/listings";
import ListingsMap from "../components/ListingsMap.vue";
import ListingFilters from "../components/ListingFilters.vue";
import ListingsSidePanel from "../components/ListingsSidePanel.vue";
import ListingsMobileList from "../components/ListingsMobileList.vue";
import MapLegend from "../components/MapLegend.vue";
import MobileViewToggle from "../components/MobileViewToggle.vue";
import { useListingFilters } from "../composables/useListingFilters";
import { useListings } from "../composables/useListings";
import { useBodyScrollLock } from "../composables/useBodyScrollLock";

const { mdAndUp } = useDisplay();
const viewMode = ref<"list" | "map">("map");
const drawerOpen = ref(true);
const teleportReady = ref(false);

const filters = useListingFilters();
const listings = useListings(filters);

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

useBodyScrollLock();

watch(
    () => [
        filters.maxPrice.value,
        filters.maxCommuteSec.value,
        filters.newWithinDays.value,
        filters.minBedrooms.value,
        filters.minBathrooms.value,
        filters.minInteriorAreaSqft.value,
    ],
    () => listings.loadInitial(),
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
});
</script>

<template>
    <Teleport to="#header-filters-slot" :disabled="!teleportReady">
        <ListingFilters :state="filters" :total="listings.total.value" />
    </Teleport>

    <Teleport to="#header-actions-slot" :disabled="!teleportReady">
        <v-btn
            v-if="mdAndUp"
            icon
            variant="text"
            size="small"
            :active="drawerOpen"
            :aria-label="drawerOpen ? 'Hide results panel' : 'Show results panel'"
            @click="drawerOpen = !drawerOpen"
        >
            <v-icon size="22">mdi-dock-right</v-icon>
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
            @update:count="mapCount = $event"
            @update:loading="mapLoading = $event"
            @error="showError"
        />
        <div
            v-if="mapLoading"
            class="map-fullbleed__loading"
            aria-live="polite"
            aria-label="Loading map listings"
        >
            <v-progress-circular indeterminate size="72" width="5" color="secondary" />
        </div>
        <ListingsSidePanel
            v-if="mdAndUp && drawerOpen"
            :items="listings.items.value"
            :total="listings.total.value"
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
        :total="listings.total.value"
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
</template>

<style scoped>
.map-fullbleed {
    position: relative;
    height: calc(100dvh - 56px);
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
</style>
