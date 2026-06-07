<script setup lang="ts">
import { ref } from "vue";
import type { Listing, SortBy, SortDir } from "../api/listings";
import type { SortOption } from "../composables/useListings";
import { useInfiniteScroll } from "../composables/useInfiniteScroll";
import ListingCard from "./ListingCard.vue";
import ListingCardSkeleton from "./ListingCardSkeleton.vue";
import ListingsSortToolbar from "./ListingsSortToolbar.vue";
import EmptyState from "./EmptyState.vue";

interface Props {
    items: Listing[];
    loading: boolean;
    hasMore: boolean;
    sortBy: SortBy;
    sortDir: SortDir;
    sortOptions: SortOption[];
    buildingTypes: number[];
}

defineProps<Props>();

const emit = defineEmits<{
    selectSort: [opt: SortOption];
    selectBuildingType: [id: number | null];
    cardClick: [item: Listing];
    loadMore: [];
}>();

const sentinelEl = ref<HTMLElement | null>(null);

useInfiniteScroll(sentinelEl, null, () => emit("loadMore"));
</script>

<template>
    <div class="listings-mobile">
        <ListingsSortToolbar
            :sort-by="sortBy"
            :sort-dir="sortDir"
            :sort-options="sortOptions"
            :building-types="buildingTypes"
            mobile
            @select-sort="emit('selectSort', $event)"
            @select-building-type="emit('selectBuildingType', $event)"
        />

        <div
            v-if="loading && items.length === 0"
            class="listing-cards listing-cards--mobile"
        >
            <ListingCardSkeleton
                v-for="i in 5"
                :key="`skeleton-mobile-${i}`"
                variant="mobile"
            />
        </div>

        <template v-else-if="items.length > 0">
            <div class="listing-cards listing-cards--mobile">
                <ListingCard
                    v-for="item in items"
                    :key="`m-${item.board}-${item.mls}`"
                    :item="item"
                    variant="mobile"
                    @click="emit('cardClick', $event)"
                />

                <div
                    v-if="hasMore && items.length > 0"
                    ref="sentinelEl"
                    class="listing-cards__sentinel"
                >
                    <v-progress-circular v-if="loading" indeterminate size="20" width="2" />
                </div>
            </div>
        </template>

        <EmptyState v-else />
    </div>
</template>

<style scoped>
.listings-mobile {
    height: calc(100dvh - 40px);
    overflow-y: auto;
    padding-bottom: calc(80px + env(safe-area-inset-bottom, 0px));
}

.listing-cards {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 10px;
}

.listing-cards--mobile {
    gap: 14px;
    padding: 10px 12px 12px;
}

.listing-cards__sentinel {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 36px;
    padding: 8px 0;
}
</style>
