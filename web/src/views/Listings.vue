<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onBeforeUnmount, watch } from "vue";
import { useDisplay } from "vuetify";
import {
    listListings,
    type Listing,
    type SortBy,
    type SortDir,
} from "../api/listings";
import ListingsMap from "../components/ListingsMap.vue";

const { mdAndUp } = useDisplay();
const viewMode = ref<"list" | "map">("map");
const drawerOpen = ref(true);

const items = ref<Listing[]>([]);
const total = ref(0);
const mapCount = ref(0);

function toggleViewMode() {
    viewMode.value = viewMode.value === "list" ? "map" : "list";
}
const limit = ref(50);
const loading = ref(false);
const error = ref<string | null>(null);
let loadGen = 0;
const sortBy = ref<SortBy>("first_seen_at");
const sortDir = ref<SortDir>("desc");

const maxPrice = ref<number | null>(null);
const maxCommuteSec = ref<number | null>(null);
const newWithinDays = ref<number | null>(null);

const priceOptions = [400000, 500000, 600000, 700000, 800000, 1000000, 1500000, 2000000];
const commuteOptions = [15, 30, 45, 60, 90];

const filterTarget = ref<"#drawer-filters-slot" | "#header-filters-slot">(
    "#header-filters-slot",
);
watch(
    [mdAndUp, drawerOpen],
    async ([md, open]) => {
        const desired = md && open ? "#drawer-filters-slot" : "#header-filters-slot";
        await nextTick();
        filterTarget.value = desired;
    },
    { immediate: true },
);

const activeFilterCount = computed(() => {
    let n = 0;
    if (maxPrice.value != null) n++;
    if (maxCommuteSec.value != null) n++;
    if (newWithinDays.value != null) n++;
    return n;
});

function formatCompactPrice(price: number): string {
    if (price >= 1_000_000) {
        const m = price / 1_000_000;
        return `$${m % 1 === 0 ? m.toFixed(0) : m.toFixed(1)}M`;
    }
    return `$${Math.round(price / 1000)}k`;
}

function applyFilters() {
    loadInitial();
}

function setMaxPrice(value: number | null) {
    maxPrice.value = value;
    applyFilters();
}

function setMaxCommute(minutes: number | null) {
    maxCommuteSec.value = minutes == null ? null : minutes * 60;
    applyFilters();
}

function toggleNewOnly() {
    newWithinDays.value = newWithinDays.value == null ? 1 : null;
    applyFilters();
}

function clearAllFilters() {
    maxPrice.value = null;
    maxCommuteSec.value = null;
    newWithinDays.value = null;
    applyFilters();
}

const hasMore = computed(() =>
    items.value.length === 0 ? true : items.value.length < total.value,
);

async function loadInitial(): Promise<void> {
    const gen = ++loadGen;
    items.value = [];
    total.value = 0;
    await loadMore(gen);
}

let inFlight: Promise<void> | null = null;

function loadMore(gen: number = loadGen): Promise<void> {
    if (gen !== loadGen) return Promise.resolve();
    if (inFlight) return inFlight.then(() => loadMore(gen));
    if (items.value.length > 0 && items.value.length >= total.value) {
        return Promise.resolve();
    }
    loading.value = true;
    error.value = null;
    inFlight = (async () => {
        try {
            const res = await listListings({
                limit: limit.value,
                offset: items.value.length,
                sortBy: sortBy.value,
                sortDir: sortDir.value,
                ...(maxPrice.value != null && { maxPrice: maxPrice.value }),
                ...(maxCommuteSec.value != null && { maxCommuteSec: maxCommuteSec.value }),
                ...(newWithinDays.value != null && { newWithinDays: newWithinDays.value }),
            });
            if (gen !== loadGen) return;
            items.value = [...items.value, ...res.items];
            total.value = res.total;
        } catch (e: any) {
            if (gen === loadGen) {
                error.value =
                    e?.response?.data?.error ?? e?.message ?? "failed to load listings";
            }
        } finally {
            if (gen === loadGen) loading.value = false;
            inFlight = null;
        }
    })();
    return inFlight;
}

function toggleSort(col: SortBy) {
    if (sortBy.value === col) {
        sortDir.value = sortDir.value === "asc" ? "desc" : "asc";
    } else {
        sortBy.value = col;
        sortDir.value = "desc";
    }
    loadInitial();
}

function sortIcon(col: SortBy): string {
    if (sortBy.value !== col) return "mdi-unfold-more-horizontal";
    return sortDir.value === "asc" ? "mdi-arrow-up" : "mdi-arrow-down";
}

const sortOptions: { value: SortBy; label: string }[] = [
    { value: "first_seen_at", label: "Newest" },
    { value: "price", label: "Price" },
    { value: "commute_time", label: "Commute" },
];

const currentSortLabel = computed(
    () => sortOptions.find((o) => o.value === sortBy.value)?.label ?? "Sort",
);

function setSort(col: SortBy, dir: SortDir) {
    if (sortBy.value === col && sortDir.value === dir) return;
    sortBy.value = col;
    sortDir.value = dir;
    loadInitial();
}

function toggleSortDir() {
    sortDir.value = sortDir.value === "asc" ? "desc" : "asc";
    loadInitial();
}

function formatPrice(price: number | null): string {
    if (price == null) return "—";
    return new Intl.NumberFormat("en-CA", {
        style: "currency",
        currency: "CAD",
        maximumFractionDigits: 0,
    }).format(price);
}

function daysSince(unix: number): number {
    const date = new Date(unix * 1000);
    const now = new Date();
    const startOfDate = new Date(date.getFullYear(), date.getMonth(), date.getDate());
    const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    return Math.round((startOfToday.getTime() - startOfDate.getTime()) / 86400000);
}

function formatDate(unix: number): string {
    const diff = daysSince(unix);
    if (diff === 0) return "today";
    if (diff === 1) return "yesterday";
    const date = new Date(unix * 1000);
    const dd = String(date.getDate()).padStart(2, "0");
    const mm = String(date.getMonth() + 1).padStart(2, "0");
    const yyyy = date.getFullYear();
    return `${dd}-${mm}-${yyyy}`;
}

function isNew(unix: number): boolean {
    return daysSince(unix) === 0;
}

function formatCommute(seconds: number | null): string {
    if (seconds == null) return "—";
    return `${Math.round(seconds / 60)} min`;
}

function commuteMapUrl(address: string | null): string | null {
    if (!address) return null;
    const params = new URLSearchParams({
        saddr: address,
        daddr: "McGill Station, Montreal, QC",
        dirflg: "r",
        ttype: "arr",
    });
    return `https://www.google.com/maps?${params.toString()}`;
}

const mapRef = ref<InstanceType<typeof ListingsMap> | null>(null);
const selectedKey = ref<string | null>(null);

function listingKey(item: Listing): string {
    return `${item.board}-${item.mls}`;
}

function openListing(item: Listing): void {
    if (!item.slug) return;
    window.open(item.slug, "_blank", "noopener,noreferrer");
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

function parseAddress(raw: string | null | undefined): {
    street: string;
    locality: string;
} {
    if (!raw) return { street: "—", locality: "" };
    const parts = raw.split("|").map((s) => s.trim()).filter(Boolean);
    const street = parts[0] ?? raw;
    const rest = parts.slice(1).join(", ");
    const locality = rest
        .replace(/\s+[A-Z]\d[A-Z]\s?\d[A-Z]\d\s*$/i, "")
        .trim();
    return { street, locality };
}

const teleportReady = ref(false);
const sidePanelBodyEl = ref<HTMLElement | null>(null);
const panelSentinelEl = ref<HTMLElement | null>(null);
const mobileSentinelEl = ref<HTMLElement | null>(null);
let panelObserver: IntersectionObserver | null = null;
let mobileObserver: IntersectionObserver | null = null;

function makeObserver(
    sentinel: HTMLElement | null,
    root: HTMLElement | null,
): IntersectionObserver | null {
    if (!sentinel) return null;
    const obs = new IntersectionObserver(
        (entries) => {
            if (entries.some((e) => e.isIntersecting)) loadMore();
        },
        { root, rootMargin: "200px" },
    );
    obs.observe(sentinel);
    return obs;
}

watch(
    [panelSentinelEl, sidePanelBodyEl],
    ([sentinel, body]) => {
        panelObserver?.disconnect();
        panelObserver = makeObserver(sentinel, body);
    },
    { flush: "post" },
);

watch(
    mobileSentinelEl,
    (sentinel) => {
        mobileObserver?.disconnect();
        mobileObserver = makeObserver(sentinel, null);
    },
    { flush: "post" },
);

onMounted(() => {
    teleportReady.value = true;
    loadInitial();
});

onBeforeUnmount(() => {
    panelObserver?.disconnect();
    mobileObserver?.disconnect();
});
</script>

<template>
    <Teleport :to="filterTarget" :disabled="!teleportReady">
        <v-menu :close-on-content-click="false" location="bottom start" offset="6">
            <template #activator="{ props }">
                <v-btn v-bind="props" rounded="pill" variant="tonal" prepend-icon="mdi-tune-variant"
                    class="filter-btn text-none" size="small">
                    Filters
                    <v-badge v-if="activeFilterCount > 0" inline color="secondary" :content="activeFilterCount"
                        class="filter-btn__badge" />
                </v-btn>
            </template>
            <v-card min-width="260" class="filter-menu">
                <div class="filter-menu__section">
                    <div class="filter-menu__title">Max price</div>
                    <div class="filter-menu__chips">
                        <v-chip size="small" :variant="maxPrice == null ? 'flat' : 'outlined'"
                            :color="maxPrice == null ? 'secondary' : undefined" @click="setMaxPrice(null)">No
                            max</v-chip>
                        <v-chip v-for="p in priceOptions" :key="p" size="small"
                            :variant="maxPrice === p ? 'flat' : 'outlined'"
                            :color="maxPrice === p ? 'secondary' : undefined"
                            @click="setMaxPrice(maxPrice === p ? null : p)">{{
                            formatCompactPrice(p) }}</v-chip>
                    </div>
                </div>
                <v-divider />
                <div class="filter-menu__section">
                    <div class="filter-menu__title">Max commute</div>
                    <div class="filter-menu__chips">
                        <v-chip size="small" :variant="maxCommuteSec == null ? 'flat' : 'outlined'"
                            :color="maxCommuteSec == null ? 'secondary' : undefined" @click="setMaxCommute(null)">No
                            max</v-chip>
                        <v-chip v-for="m in commuteOptions" :key="m" size="small"
                            :variant="maxCommuteSec === m * 60 ? 'flat' : 'outlined'"
                            :color="maxCommuteSec === m * 60 ? 'secondary' : undefined"
                            @click="setMaxCommute(maxCommuteSec === m * 60 ? null : m)">{{ m }} min</v-chip>
                    </div>
                </div>
                <v-divider />
                <div class="filter-menu__section">
                    <div class="filter-menu__title">Recency</div>
                    <div class="filter-menu__chips">
                        <v-chip size="small" :variant="newWithinDays != null ? 'flat' : 'outlined'"
                            :color="newWithinDays != null ? 'secondary' : undefined" @click="toggleNewOnly">New today
                            only</v-chip>
                    </div>
                </div>
                <template v-if="activeFilterCount > 0">
                    <v-divider />
                    <div class="filter-menu__footer">
                        <v-btn variant="text" size="small" class="text-none" @click="clearAllFilters">Clear all</v-btn>
                    </div>
                </template>
            </v-card>
        </v-menu>

        <v-menu v-if="maxPrice != null" location="bottom start" offset="6">
            <template #activator="{ props }">
                <v-chip v-bind="props" color="secondary" variant="tonal" class="filter-chip" size="small" closable
                    @click:close.stop="setMaxPrice(null)">
                    ≤ {{ formatCompactPrice(maxPrice) }}
                </v-chip>
            </template>
            <v-list density="compact">
                <v-list-item :active="maxPrice == null" title="No max" @click="setMaxPrice(null)" />
                <v-divider />
                <v-list-item v-for="p in priceOptions" :key="p" :active="maxPrice === p"
                    :title="`Up to ${formatCompactPrice(p)}`" @click="setMaxPrice(p)" />
            </v-list>
        </v-menu>

        <v-menu v-if="maxCommuteSec != null" location="bottom start" offset="6">
            <template #activator="{ props }">
                <v-chip v-bind="props" color="secondary" variant="tonal" class="filter-chip" size="small" closable
                    @click:close.stop="setMaxCommute(null)">
                    ≤ {{ Math.round(maxCommuteSec / 60) }} min
                </v-chip>
            </template>
            <v-list density="compact">
                <v-list-item :active="maxCommuteSec == null" title="No max" @click="setMaxCommute(null)" />
                <v-divider />
                <v-list-item v-for="m in commuteOptions" :key="m" :active="maxCommuteSec === m * 60"
                    :title="`Up to ${m} min`" @click="setMaxCommute(m)" />
            </v-list>
        </v-menu>

        <v-chip v-if="newWithinDays != null" color="secondary" variant="tonal" class="filter-chip" size="small" closable
            @click:close.stop="toggleNewOnly">
            New today
        </v-chip>
    </Teleport>

    <Teleport to="#header-actions-slot" :disabled="!teleportReady">
        <v-btn v-if="mdAndUp" icon variant="text" size="small" :active="drawerOpen"
            :aria-label="drawerOpen ? 'Hide results panel' : 'Show results panel'" @click="drawerOpen = !drawerOpen">
            <v-icon size="22">mdi-dock-right</v-icon>
        </v-btn>
    </Teleport>

    <template v-if="mdAndUp || viewMode === 'map'">
        <div class="map-fullbleed" :class="{ 'map-fullbleed--with-panel': mdAndUp && drawerOpen }">
            <ListingsMap ref="mapRef" class="map-fullbleed__map" :max-price="maxPrice" :max-commute-sec="maxCommuteSec"
                :new-within-days="newWithinDays" @update:count="mapCount = $event" />
            <aside v-if="mdAndUp && drawerOpen" class="listings-side-panel">
                <div id="drawer-filters-slot" class="listings-side-panel__filters" />
                <div ref="sidePanelBodyEl" class="listings-side-panel__body">
                    <v-alert v-if="error" type="error" variant="tonal" class="ma-3">{{ error }}</v-alert>

                    <div v-if="loading && items.length === 0" class="text-center py-8">
                        <v-progress-circular indeterminate />
                    </div>

                    <template v-else-if="items.length > 0">
                    <div class="listings-sort">
                        <span class="listings-sort__count">
                            {{ total.toLocaleString() }} listing{{ total === 1 ? "" : "s" }}
                        </span>
                        <v-menu location="bottom end" offset="4">
                            <template #activator="{ props }">
                                <v-btn v-bind="props" variant="text" size="small" density="comfortable"
                                    class="listings-sort__btn text-none" append-icon="mdi-menu-down">
                                    <v-icon size="small" start>{{
                                        sortDir === "asc" ? "mdi-arrow-up" : "mdi-arrow-down"
                                    }}</v-icon>
                                    {{ currentSortLabel }}
                                </v-btn>
                            </template>
                            <v-list density="compact" min-width="180">
                                <v-list-item v-for="opt in sortOptions" :key="opt.value"
                                    :active="sortBy === opt.value" :title="opt.label"
                                    @click="setSort(opt.value, 'desc')" />
                                <v-divider />
                                <v-list-item :title="sortDir === 'asc' ? 'Descending' : 'Ascending'"
                                    :prepend-icon="sortDir === 'asc' ? 'mdi-arrow-down' : 'mdi-arrow-up'"
                                    @click="toggleSortDir" />
                            </v-list>
                        </v-menu>
                    </div>

                    <div class="listing-cards listing-cards--panel">
                        <div v-for="item in items" :key="`p-${item.board}-${item.mls}`" role="button" tabindex="0"
                            class="listing-card listing-card--interactive"
                            :class="{ 'listing-card--selected': selectedKey === listingKey(item) }"
                            @click="focusListingOnMap(item)" @keydown.enter.prevent="focusListingOnMap(item)"
                            @keydown.space.prevent="focusListingOnMap(item)" @mouseenter="highlightListingOnMap(item)"
                            @mouseleave="clearMapHighlight" @focus="highlightListingOnMap(item)"
                            @blur="clearMapHighlight">
                            <div class="listing-card__top">
                                <span class="listing-card__price">{{
                                    formatPrice(item.currentPrice)
                                }}</span>
                                <v-chip v-if="isNew(item.firstSeenAt)" size="x-small" color="secondary" variant="flat"
                                    class="listing-card__new">new</v-chip>
                            </div>
                            <div class="listing-card__street">
                                {{ parseAddress(item.address).street }}
                            </div>
                            <div v-if="parseAddress(item.address).locality" class="listing-card__locality">
                                {{ parseAddress(item.address).locality }}
                            </div>
                            <div class="listing-card__meta">
                                <a v-if="item.commuteSecondsDowntown != null && item.address"
                                    :href="commuteMapUrl(item.address) ?? '#'" target="_blank" rel="noopener noreferrer"
                                    class="listing-card__commute listing-card__commute--link" @click.stop>
                                    <v-icon size="small">mdi-train</v-icon>
                                    <span>{{ formatCommute(item.commuteSecondsDowntown) }}</span>
                                    <span class="listing-card__commute-label">to downtown</span>
                                    <v-icon size="x-small"
                                        class="listing-card__commute-chevron">mdi-chevron-right</v-icon>
                                </a>
                                <span v-else class="listing-card__commute listing-card__commute--muted">
                                    <v-icon size="small">mdi-train</v-icon>
                                    —
                                </span>
                                <span class="listing-card__seen">{{
                                    formatDate(item.firstSeenAt)
                                }}</span>
                            </div>
                        </div>
                    </div>
                    </template>

                    <div v-else class="text-medium-emphasis text-center py-8">
                        No listings found.
                    </div>

                    <div v-if="hasMore && items.length > 0" ref="panelSentinelEl" class="listings-side-panel__sentinel">
                        <v-progress-circular v-if="loading" indeterminate size="20" width="2" />
                    </div>
                </div>
            </aside>
        </div>
    </template>

    <v-container v-if="!mdAndUp && viewMode === 'list'" fluid class="pa-2 pa-sm-6 listings-container">
        <v-card>
            <v-alert v-if="error" type="error" variant="tonal" class="ma-3">{{
                error
            }}</v-alert>

            <v-card-text v-if="loading && items.length === 0" class="text-center py-8">
                <v-progress-circular indeterminate />
            </v-card-text>

            <template v-else-if="items.length > 0">
                <v-table density="comfortable" class="listings-table d-none d-md-table">
                    <thead>
                        <tr>
                            <th>Address</th>
                            <th class="sortable-col text-right" @click="toggleSort('price')">
                                Price
                                <v-icon size="small" class="sort-icon">{{
                                    sortIcon("price")
                                }}</v-icon>
                            </th>
                            <th class="sortable-col" @click="toggleSort('first_seen_at')">
                                First Seen
                                <v-icon size="small" class="sort-icon">{{
                                    sortIcon("first_seen_at")
                                }}</v-icon>
                            </th>
                            <th class="sortable-col" @click="toggleSort('commute_time')">
                                Commute Time
                                <v-icon size="small" class="sort-icon">{{
                                    sortIcon("commute_time")
                                }}</v-icon>
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="item in items" :key="`${item.board}-${item.mls}`" class="listing-row">
                            <td>
                                <v-tooltip v-if="item.slug && item.address" location="top" open-delay="400">
                                    <template #activator="{ props }">
                                        <a v-bind="props" :href="item.slug" target="_blank" rel="noopener noreferrer"
                                            class="address-link">
                                            <span>{{ item.address }}</span>
                                            <v-icon size="small" class="address-link__icon">mdi-open-in-new</v-icon>
                                        </a>
                                    </template>
                                    <span>View on Realtor.ca</span>
                                </v-tooltip>
                                <template v-else>{{ item.address || "—" }}</template>
                            </td>
                            <td class="text-right">
                                {{ formatPrice(item.currentPrice) }}
                            </td>
                            <td>
                                <span class="first-seen">
                                    <span>{{ formatDate(item.firstSeenAt) }}</span>
                                    <v-chip v-if="isNew(item.firstSeenAt)" size="x-small" color="secondary"
                                        variant="outlined">new</v-chip>
                                </span>
                            </td>
                            <td>
                                <v-tooltip v-if="
                                    item.commuteSecondsDowntown != null &&
                                    item.address
                                " location="top" open-delay="400">
                                    <template #activator="{ props }">
                                        <a v-bind="props" :href="commuteMapUrl(item.address) ?? '#'" target="_blank"
                                            rel="noopener noreferrer" class="commute-link">
                                            <v-icon size="small" class="commute-link__icon">mdi-directions</v-icon>
                                            <span>{{ formatCommute(item.commuteSecondsDowntown) }}</span>
                                        </a>
                                    </template>
                                    <span>Get directions to downtown</span>
                                </v-tooltip>
                                <span v-else class="text-medium-emphasis">
                                    {{ formatCommute(item.commuteSecondsDowntown) }}
                                </span>
                            </td>
                        </tr>
                    </tbody>
                </v-table>

                <div class="listings-sort listings-sort--mobile d-md-none">
                    <span class="listings-sort__count">
                        {{ total.toLocaleString() }} listing{{ total === 1 ? "" : "s" }}
                    </span>
                    <v-menu location="bottom end" offset="4">
                        <template #activator="{ props }">
                            <v-btn v-bind="props" variant="text" size="small" density="comfortable"
                                class="listings-sort__btn text-none" append-icon="mdi-menu-down">
                                <v-icon size="small" start>{{
                                    sortDir === "asc" ? "mdi-arrow-up" : "mdi-arrow-down"
                                }}</v-icon>
                                {{ currentSortLabel }}
                            </v-btn>
                        </template>
                        <v-list density="compact" min-width="180">
                            <v-list-item v-for="opt in sortOptions" :key="opt.value"
                                :active="sortBy === opt.value" :title="opt.label"
                                @click="setSort(opt.value, 'desc')" />
                            <v-divider />
                            <v-list-item :title="sortDir === 'asc' ? 'Descending' : 'Ascending'"
                                :prepend-icon="sortDir === 'asc' ? 'mdi-arrow-down' : 'mdi-arrow-up'"
                                @click="toggleSortDir" />
                        </v-list>
                    </v-menu>
                </div>

                <div class="d-md-none listing-cards listing-cards--mobile">
                    <div v-for="item in items" :key="`m-${item.board}-${item.mls}`" role="link" tabindex="0"
                        class="listing-card" @click="openListing(item)" @keydown.enter.prevent="openListing(item)"
                        @keydown.space.prevent="openListing(item)">
                        <div class="listing-card__top">
                            <span class="listing-card__price">{{
                                formatPrice(item.currentPrice)
                            }}</span>
                            <v-chip v-if="isNew(item.firstSeenAt)" size="x-small" color="secondary" variant="flat"
                                class="listing-card__new">new</v-chip>
                        </div>
                        <div class="listing-card__street">
                            {{ parseAddress(item.address).street }}
                        </div>
                        <div v-if="parseAddress(item.address).locality" class="listing-card__locality">
                            {{ parseAddress(item.address).locality }}
                        </div>
                        <div class="listing-card__meta">
                            <a v-if="item.commuteSecondsDowntown != null && item.address"
                                :href="commuteMapUrl(item.address) ?? '#'" target="_blank" rel="noopener noreferrer"
                                class="listing-card__commute listing-card__commute--link" @click.stop>
                                <v-icon size="small">mdi-train</v-icon>
                                <span>{{ formatCommute(item.commuteSecondsDowntown) }}</span>
                                <span class="listing-card__commute-label">to downtown</span>
                                <v-icon size="x-small" class="listing-card__commute-chevron">mdi-chevron-right</v-icon>
                            </a>
                            <span v-else class="listing-card__commute listing-card__commute--muted">
                                <v-icon size="small">mdi-train</v-icon>
                                —
                            </span>
                            <span class="listing-card__seen">{{
                                formatDate(item.firstSeenAt)
                            }}</span>
                        </div>
                    </div>

                    <div v-if="hasMore && items.length > 0" ref="mobileSentinelEl" class="listing-cards__sentinel">
                        <v-progress-circular v-if="loading" indeterminate size="20" width="2" />
                    </div>
                </div>
            </template>

            <v-card-text v-else class="text-medium-emphasis text-center py-8">
                No listings found.
            </v-card-text>
        </v-card>
    </v-container>

    <v-btn v-if="!mdAndUp" class="view-switch-pill text-none" color="secondary" variant="flat" rounded="pill"
        size="large" elevation="8" @click="toggleViewMode">
        <v-icon start>{{ viewMode === "list" ? "mdi-map" : "mdi-format-list-bulleted" }}</v-icon>
        {{ viewMode === "list" ? "Show map" : "Show list" }}
    </v-btn>
</template>

<style scoped>
.filter-btn {
    letter-spacing: normal;
    font-weight: 500;
}

.filter-btn__badge {
    margin-inline-start: 8px;
}

.filter-chip {
    font-weight: 500;
}

@media (max-width: 959.98px) {
    .filter-chip {
        display: none;
    }
}

.filter-menu {
    padding: 4px 0;
}

.filter-menu__section {
    padding: 12px 14px;
}

.filter-menu__title {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: rgba(var(--v-theme-on-surface), 0.6);
    margin-bottom: 8px;
}

.filter-menu__chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
}

.filter-menu__footer {
    display: flex;
    justify-content: flex-end;
    padding: 6px 8px;
}

.listings-container {
    padding-bottom: calc(96px + env(safe-area-inset-bottom, 0px)) !important;
}

.map-fullbleed {
    height: calc(100dvh - 56px);
    width: 100%;
    display: flex;
}

.map-fullbleed__map {
    flex: 1 1 auto;
    min-height: 0;
    min-width: 0;
}

.listings-side-panel {
    flex: 0 0 360px;
    display: flex;
    flex-direction: column;
    min-height: 0;
    border-left: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    background-color: rgb(var(--v-theme-surface));
}

.listings-side-panel__filters {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    padding: 10px 12px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    flex: 0 0 auto;
}

.listings-side-panel__filters:empty {
    display: none;
}

.listings-side-panel__body {
    flex: 1 1 auto;
    overflow-y: auto;
    min-height: 0;
}

.listings-side-panel__sentinel,
.listing-cards__sentinel {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 36px;
    padding: 8px 0;
}

.listing-cards--panel {
    padding: 10px;
}

.listings-sort {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 6px 10px 6px 14px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.06);
}

.listings-sort--mobile {
    padding: 6px 10px 6px 14px;
    border-bottom: 0;
}

.listings-sort__count {
    font-size: 0.8125rem;
    font-weight: 500;
    color: rgba(var(--v-theme-on-surface), 0.75);
}

.listings-sort__btn {
    letter-spacing: normal;
    font-weight: 500;
}

.view-switch-pill {
    position: fixed;
    left: calc(50% + (var(--v-layout-left, 0px) - var(--v-layout-right, 0px)) / 2);
    bottom: calc(36px + env(safe-area-inset-bottom, 0px));
    transform: translateX(-50%);
    z-index: 1000;
    letter-spacing: normal;
    font-weight: 600;
    padding-inline: 22px;
}

.sortable-col {
    cursor: pointer;
    user-select: none;
    white-space: nowrap;
}

.sortable-col:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.04);
}

.sort-icon {
    opacity: 0.5;
    vertical-align: middle;
}

.listings-table :deep(tbody tr.listing-row) {
    transition: background-color 120ms ease;
}

.listings-table :deep(tbody tr.listing-row:hover) {
    background-color: rgba(var(--v-theme-on-surface), 0.035);
}

.address-link {
    color: inherit;
    text-decoration: none;
    display: inline-flex;
    align-items: center;
    gap: 6px;
}

.address-link__icon {
    opacity: 0;
    transform: translateX(-4px);
    transition: opacity 120ms ease, transform 120ms ease;
    color: rgb(var(--v-theme-secondary));
}

.listing-row:hover .address-link__icon {
    opacity: 0.75;
    transform: translateX(0);
}

.address-link:hover {
    color: rgb(var(--v-theme-secondary));
    text-decoration: underline;
}

.address-link:hover .address-link__icon {
    opacity: 1;
}

.first-seen {
    display: inline-flex;
    align-items: center;
    gap: 8px;
}

.commute-link {
    color: rgb(var(--v-theme-secondary));
    text-decoration: none;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
}

.commute-link__icon {
    opacity: 0.7;
    transition: opacity 120ms ease;
}

.commute-link:hover {
    text-decoration: underline;
}

.commute-link:hover .commute-link__icon {
    opacity: 1;
}

.listing-cards {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 10px;
}

.listing-card {
    position: relative;
    display: block;
    text-decoration: none;
    color: inherit;
    background-color: rgba(var(--v-theme-on-surface), 0.03);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    border-radius: 10px;
    padding: 14px 14px 12px;
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease;
}

.listing-card:active {
    background-color: rgba(var(--v-theme-on-surface), 0.06);
}

.listing-card:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: 2px;
}

.listing-card--interactive:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.06);
    border-color: rgba(var(--v-theme-on-surface), 0.18);
}

.listing-card--selected,
.listing-card--selected:hover {
    background-color: rgba(var(--v-theme-secondary), 0.12);
    border-color: rgba(var(--v-theme-secondary), 0.55);
    box-shadow: inset 0 0 0 1px rgba(var(--v-theme-secondary), 0.45);
}

.listing-card__top {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
    padding-right: 56px;
}

.listing-card__price {
    font-size: 1.25rem;
    font-weight: 600;
    line-height: 1.2;
}

.listing-card__new {
    position: absolute;
    top: 12px;
    right: 12px;
    margin-left: 0;
}

.listing-card__street {
    font-size: 0.95rem;
    line-height: 1.3;
}

.listing-card__locality {
    font-size: 0.8125rem;
    color: rgba(var(--v-theme-on-surface), 0.65);
    margin-top: 2px;
}

.listing-card__meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: 12px;
    padding-top: 10px;
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    font-size: 0.8125rem;
}

.listing-card__commute {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: rgb(var(--v-theme-secondary));
    font-weight: 500;
}

.listing-card__commute--link {
    text-decoration: none;
    padding: 6px 10px 6px 8px;
    margin: -6px 0 -6px -8px;
    border-radius: 999px;
    background-color: rgba(var(--v-theme-secondary), 0.1);
    border: 1px solid rgba(var(--v-theme-secondary), 0.25);
    transition: background-color 120ms ease, border-color 120ms ease;
}

.listing-card__commute--link:active {
    background-color: rgba(var(--v-theme-secondary), 0.2);
}

.listing-card__commute--link:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: 2px;
}

.listing-card__commute-chevron {
    opacity: 0.6;
    margin-left: 2px;
}

.listing-card__commute--muted {
    color: rgba(var(--v-theme-on-surface), 0.5);
    font-weight: 400;
}

.listing-card__commute-label {
    color: rgba(var(--v-theme-on-surface), 0.55);
    font-weight: 400;
    margin-left: 2px;
}

.listing-card__seen {
    margin-left: auto;
    color: rgba(var(--v-theme-on-surface), 0.45);
    font-size: 0.75rem;
}
</style>
