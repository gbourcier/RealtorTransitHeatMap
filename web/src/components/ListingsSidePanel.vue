<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import type { Listing, SortBy, SortDir } from "../api/listings";
import type { SortOption } from "../composables/useListings";
import { useInfiniteScroll } from "../composables/useInfiniteScroll";
import ListingCard from "./ListingCard.vue";
import ListingCardSkeleton from "./ListingCardSkeleton.vue";
import ListingsSortToolbar from "./ListingsSortToolbar.vue";
import EmptyState from "./EmptyState.vue";

const SKELETON_ROW_HEIGHT = 156;
const SKELETON_LIST_PADDING = 20;
const MIN_PANEL_WIDTH = 320;
const MAX_PANEL_WIDTH = 680;
const MIN_MAP_WIDTH = 360;
const RESIZE_STEP = 24;

interface Props {
    items: Listing[];
    loading: boolean;
    hasMore: boolean;
    sortBy: SortBy;
    sortDir: SortDir;
    sortOptions: SortOption[];
    selectedKey: string | null;
    width: number;
}

const props = defineProps<Props>();

const emit = defineEmits<{
    selectSort: [opt: SortOption];
    cardClick: [item: Listing];
    cardHover: [item: Listing];
    cardLeave: [];
    loadMore: [];
    "update:width": [width: number];
}>();

const bodyEl = ref<HTMLElement | null>(null);
const sentinelEl = ref<HTMLElement | null>(null);
const bodyHeight = ref(0);
const isResizing = ref(false);

let resizeObserver: ResizeObserver | null = null;
let startX = 0;
let startWidth = 0;

onMounted(() => {
    if (!bodyEl.value) return;
    syncClampedWidth();
    bodyHeight.value = bodyEl.value.clientHeight;
    resizeObserver = new ResizeObserver((entries) => {
        for (const entry of entries) {
            bodyHeight.value = entry.contentRect.height;
        }
    });
    resizeObserver.observe(bodyEl.value);
    window.addEventListener("resize", syncClampedWidth);
});

onBeforeUnmount(() => {
    resizeObserver?.disconnect();
    resizeObserver = null;
    window.removeEventListener("resize", syncClampedWidth);
    stopResize();
});

const skeletonCount = computed(() => {
    const available = Math.max(0, bodyHeight.value - SKELETON_LIST_PADDING);
    return Math.max(1, Math.floor(available / SKELETON_ROW_HEIGHT));
});

useInfiniteScroll(sentinelEl, bodyEl, () => emit("loadMore"));

function listingKey(item: Listing): string {
    return `${item.board}-${item.mls}`;
}

const clampedWidth = computed(() => clampPanelWidth(props.width));

function maxPanelWidth(): number {
    if (typeof window === "undefined") return MAX_PANEL_WIDTH;
    return Math.max(MIN_PANEL_WIDTH, Math.min(MAX_PANEL_WIDTH, window.innerWidth - MIN_MAP_WIDTH));
}

function clampPanelWidth(width: number): number {
    return Math.min(maxPanelWidth(), Math.max(MIN_PANEL_WIDTH, Math.round(width)));
}

function setPanelWidth(width: number): void {
    emit("update:width", clampPanelWidth(width));
}

function syncClampedWidth(): void {
    const width = clampPanelWidth(props.width);
    if (width !== props.width) emit("update:width", width);
}

function onResizePointerDown(event: PointerEvent): void {
    if (event.button !== 0) return;
    event.preventDefault();
    startX = event.clientX;
    startWidth = clampedWidth.value;
    isResizing.value = true;
    document.body.classList.add("listings-side-panel-resizing");
    window.addEventListener("pointermove", onResizePointerMove);
    window.addEventListener("pointerup", stopResize);
    window.addEventListener("pointercancel", stopResize);
}

function onResizePointerMove(event: PointerEvent): void {
    setPanelWidth(startWidth + startX - event.clientX);
}

function stopResize(): void {
    if (!isResizing.value) return;
    isResizing.value = false;
    document.body.classList.remove("listings-side-panel-resizing");
    window.removeEventListener("pointermove", onResizePointerMove);
    window.removeEventListener("pointerup", stopResize);
    window.removeEventListener("pointercancel", stopResize);
}

function onResizeKeydown(event: KeyboardEvent): void {
    if (event.key === "ArrowLeft") {
        event.preventDefault();
        setPanelWidth(clampedWidth.value + RESIZE_STEP);
    } else if (event.key === "ArrowRight") {
        event.preventDefault();
        setPanelWidth(clampedWidth.value - RESIZE_STEP);
    } else if (event.key === "Home") {
        event.preventDefault();
        setPanelWidth(MIN_PANEL_WIDTH);
    } else if (event.key === "End") {
        event.preventDefault();
        setPanelWidth(maxPanelWidth());
    }
}
</script>

<template>
    <aside class="listings-side-panel" :style="{ '--listings-panel-width': `${clampedWidth}px` }">
        <div
            class="listings-side-panel__resize"
            :class="{ 'listings-side-panel__resize--active': isResizing }"
            role="separator"
            aria-label="Resize results panel"
            aria-orientation="vertical"
            :aria-valuemin="MIN_PANEL_WIDTH"
            :aria-valuemax="maxPanelWidth()"
            :aria-valuenow="clampedWidth"
            tabindex="0"
            @pointerdown="onResizePointerDown"
            @keydown="onResizeKeydown"
        />
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

            <EmptyState v-else />

            <div v-if="hasMore && items.length > 0" ref="sentinelEl" class="listings-side-panel__sentinel">
                <v-progress-circular v-if="loading" indeterminate size="20" width="2" />
            </div>
        </div>
    </aside>
</template>

<style scoped>
.listings-side-panel {
    position: relative;
    flex: 0 0 var(--listings-panel-width);
    width: var(--listings-panel-width);
    display: flex;
    flex-direction: column;
    min-height: 0;
    border-left: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    background-color: rgb(var(--v-theme-surface));
}

.listings-side-panel__resize {
    position: absolute;
    top: 0;
    bottom: 0;
    left: -5px;
    z-index: 10;
    width: 10px;
    cursor: ew-resize;
    touch-action: none;
}

.listings-side-panel__resize::before {
    content: "";
    position: absolute;
    top: 0;
    bottom: 0;
    left: 4px;
    width: 2px;
    background-color: transparent;
    transition: background-color 120ms ease, box-shadow 120ms ease;
}

.listings-side-panel__resize:hover::before,
.listings-side-panel__resize:focus-visible::before,
.listings-side-panel__resize--active::before {
    background-color: rgb(var(--v-theme-primary));
    box-shadow: 0 0 0 1px rgba(var(--v-theme-primary), 0.2);
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

<style>
.listings-side-panel-resizing {
    cursor: ew-resize;
    user-select: none;
}
</style>
