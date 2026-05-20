<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import type { Listing, SortBy, SortDir } from "../api/listings";
import type { SortOption } from "../composables/useListings";
import { useInfiniteScroll } from "../composables/useInfiniteScroll";
import ListingCard from "./ListingCard.vue";
import ListingCardSkeleton from "./ListingCardSkeleton.vue";
import ListingsSortToolbar from "./ListingsSortToolbar.vue";

const SKELETON_ROW_HEIGHT = 156;
const SKELETON_LIST_PADDING = 20;

interface Props {
    items: Listing[];
    loading: boolean;
    hasMore: boolean;
    sortBy: SortBy;
    sortDir: SortDir;
    sortOptions: SortOption[];
    selectedKey: string | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
    selectSort: [opt: SortOption];
    cardClick: [item: Listing];
    cardHover: [item: Listing];
    cardLeave: [];
    loadMore: [];
}>();

const bodyEl = ref<HTMLElement | null>(null);
const sentinelEl = ref<HTMLElement | null>(null);
const bodyHeight = ref(0);

let resizeObserver: ResizeObserver | null = null;

onMounted(() => {
    if (!bodyEl.value) return;
    bodyHeight.value = bodyEl.value.clientHeight;
    resizeObserver = new ResizeObserver((entries) => {
        for (const entry of entries) {
            bodyHeight.value = entry.contentRect.height;
        }
    });
    resizeObserver.observe(bodyEl.value);
});

onBeforeUnmount(() => {
    resizeObserver?.disconnect();
    resizeObserver = null;
});

const skeletonCount = computed(() => {
    const available = Math.max(0, bodyHeight.value - SKELETON_LIST_PADDING);
    return Math.max(1, Math.floor(available / SKELETON_ROW_HEIGHT));
});

useInfiniteScroll(sentinelEl, bodyEl, () => emit("loadMore"));

function listingKey(item: Listing): string {
    return `${item.board}-${item.mls}`;
}
</script>

<template>
    <aside class="listings-side-panel">
        <ListingsSortToolbar :sort-by="sortBy" :sort-dir="sortDir" :sort-options="sortOptions"
            @select-sort="emit('selectSort', $event)" />
        <div ref="bodyEl" class="listings-side-panel__body">
            <div v-if="loading && items.length === 0" class="listing-cards listing-cards--panel">
                <ListingCardSkeleton v-for="i in skeletonCount" :key="`skeleton-panel-${i}`" variant="panel" />
            </div>

            <template v-else-if="items.length > 0">
                <div class="listing-cards listing-cards--panel">
                    <ListingCard v-for="item in items" :key="`p-${item.board}-${item.mls}`" :item="item" variant="panel"
                        :selected="props.selectedKey === listingKey(item)" @click="emit('cardClick', $event)"
                        @hover="emit('cardHover', $event)" @leave="emit('cardLeave')" />
                </div>
            </template>

            <div v-else class="text-medium-emphasis text-center py-8">
                No listings found.
            </div>

            <div v-if="hasMore && items.length > 0" ref="sentinelEl" class="listings-side-panel__sentinel">
                <v-progress-circular v-if="loading" indeterminate size="20" width="2" />
            </div>
        </div>
    </aside>
</template>

<style scoped>
.listings-side-panel {
    flex: 0 0 360px;
    display: flex;
    flex-direction: column;
    min-height: 0;
    border-left: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    background-color: rgb(var(--v-theme-surface));
}

.listings-side-panel__body {
    flex: 1 1 auto;
    overflow-y: auto;
    min-height: 0;
}

.listings-side-panel__sentinel {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 36px;
    padding: 8px 0;
}

.listing-cards {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 10px;
}

.listing-cards--panel {
    padding: 10px;
}
</style>
