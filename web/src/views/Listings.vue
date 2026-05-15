<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import {
    listListings,
    type Listing,
    type SortBy,
    type SortDir,
} from "../api/listings";

const items = ref<Listing[]>([]);
const total = ref(0);
const limit = ref(50);
const page = ref(1);
const loading = ref(false);
const error = ref<string | null>(null);
const sortBy = ref<SortBy>("first_seen_at");
const sortDir = ref<SortDir>("desc");

const maxPrice = ref<number | null>(null);
const maxCommuteSec = ref<number | null>(null);
const newWithinDays = ref<number | null>(null);

const priceOptions = [400000, 500000, 600000, 700000, 800000, 1000000, 1500000, 2000000];
const commuteOptions = [15, 30, 45, 60, 90];

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

function priceChipLabel(): string {
    return maxPrice.value == null
        ? "Max price"
        : `≤ ${formatCompactPrice(maxPrice.value)}`;
}

function commuteChipLabel(): string {
    return maxCommuteSec.value == null
        ? "Max commute"
        : `≤ ${Math.round(maxCommuteSec.value / 60)} min`;
}

function applyFilters() {
    page.value = 1;
    load();
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
    newWithinDays.value = newWithinDays.value == null ? 7 : null;
    applyFilters();
}

function clearAllFilters() {
    maxPrice.value = null;
    maxCommuteSec.value = null;
    newWithinDays.value = null;
    applyFilters();
}

const offset = computed(() => (page.value - 1) * limit.value);
const pageCount = computed(() => Math.ceil(total.value / limit.value));

async function load() {
    loading.value = true;
    error.value = null;
    try {
        const res = await listListings({
            limit: limit.value,
            offset: offset.value,
            sortBy: sortBy.value,
            sortDir: sortDir.value,
            ...(maxPrice.value != null && { maxPrice: maxPrice.value }),
            ...(maxCommuteSec.value != null && { maxCommuteSec: maxCommuteSec.value }),
            ...(newWithinDays.value != null && { newWithinDays: newWithinDays.value }),
        });
        items.value = res.items;
        total.value = res.total;
    } catch (e: any) {
        error.value =
            e?.response?.data?.error ?? e?.message ?? "failed to load listings";
    } finally {
        loading.value = false;
    }
}

function toggleSort(col: SortBy) {
    if (sortBy.value === col) {
        sortDir.value = sortDir.value === "asc" ? "desc" : "asc";
    } else {
        sortBy.value = col;
        sortDir.value = "desc";
    }
    page.value = 1;
    load();
}

function sortIcon(col: SortBy): string {
    if (sortBy.value !== col) return "mdi-unfold-more-horizontal";
    return sortDir.value === "asc" ? "mdi-arrow-up" : "mdi-arrow-down";
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

function openListing(item: Listing): void {
    if (!item.slug) return;
    window.open(item.slug, "_blank", "noopener,noreferrer");
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

onMounted(load);
</script>

<template>
    <v-container fluid class="pa-2 pa-sm-6">
        <v-card>
            <v-card-title class="d-flex align-center">
                <span>Listings</span>
                <v-spacer />
                <span v-if="total > 0" class="text-body-2 text-medium-emphasis">{{ total }} total</span>
            </v-card-title>
            <v-divider />

            <div class="filter-bar px-3 py-2">
                <v-menu>
                    <template #activator="{ props }">
                        <v-btn
                            v-bind="props"
                            :variant="maxPrice != null ? 'tonal' : 'text'"
                            :color="maxPrice != null ? 'secondary' : undefined"
                            rounded="pill"
                            append-icon="mdi-chevron-down"
                            class="filter-btn text-none"
                        >{{ priceChipLabel() }}</v-btn>
                    </template>
                    <v-list density="compact">
                        <v-list-item
                            :active="maxPrice == null"
                            title="No max"
                            @click="setMaxPrice(null)"
                        />
                        <v-divider />
                        <v-list-item
                            v-for="p in priceOptions"
                            :key="p"
                            :active="maxPrice === p"
                            :title="`Up to ${formatCompactPrice(p)}`"
                            @click="setMaxPrice(p)"
                        />
                    </v-list>
                </v-menu>

                <v-menu>
                    <template #activator="{ props }">
                        <v-btn
                            v-bind="props"
                            :variant="maxCommuteSec != null ? 'tonal' : 'text'"
                            :color="maxCommuteSec != null ? 'secondary' : undefined"
                            rounded="pill"
                            append-icon="mdi-chevron-down"
                            class="filter-btn text-none"
                        >{{ commuteChipLabel() }}</v-btn>
                    </template>
                    <v-list density="compact">
                        <v-list-item
                            :active="maxCommuteSec == null"
                            title="No max"
                            @click="setMaxCommute(null)"
                        />
                        <v-divider />
                        <v-list-item
                            v-for="m in commuteOptions"
                            :key="m"
                            :active="maxCommuteSec === m * 60"
                            :title="`Up to ${m} min`"
                            @click="setMaxCommute(m)"
                        />
                    </v-list>
                </v-menu>

                <v-btn
                    :variant="newWithinDays != null ? 'tonal' : 'text'"
                    :color="newWithinDays != null ? 'secondary' : undefined"
                    rounded="pill"
                    class="filter-btn text-none"
                    @click="toggleNewOnly"
                >New only</v-btn>

                <v-btn
                    v-if="activeFilterCount > 0"
                    variant="text"
                    size="small"
                    class="filter-clear text-none"
                    @click="clearAllFilters"
                >Clear</v-btn>
            </div>

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

            <div class="d-md-none listing-cards">
                <div
                    v-for="item in items"
                    :key="`m-${item.board}-${item.mls}`"
                    role="link"
                    tabindex="0"
                    class="listing-card"
                    @click="openListing(item)"
                    @keydown.enter.prevent="openListing(item)"
                    @keydown.space.prevent="openListing(item)"
                >
                    <div class="listing-card__top">
                        <span class="listing-card__price">{{
                            formatPrice(item.currentPrice)
                        }}</span>
                        <v-chip
                            v-if="isNew(item.firstSeenAt)"
                            size="x-small"
                            color="secondary"
                            variant="flat"
                            class="listing-card__new"
                        >new</v-chip>
                    </div>
                    <div class="listing-card__street">
                        {{ parseAddress(item.address).street }}
                    </div>
                    <div
                        v-if="parseAddress(item.address).locality"
                        class="listing-card__locality"
                    >
                        {{ parseAddress(item.address).locality }}
                    </div>
                    <div class="listing-card__meta">
                        <a
                            v-if="item.commuteSecondsDowntown != null && item.address"
                            :href="commuteMapUrl(item.address) ?? '#'"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="listing-card__commute listing-card__commute--link"
                            @click.stop
                        >
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
            </div>
            </template>

            <v-card-text v-else class="text-medium-emphasis text-center py-8">
                No listings found.
            </v-card-text>

            <template v-if="pageCount > 1">
                <v-divider />
                <v-card-text class="d-flex justify-center pa-3">
                    <v-pagination :model-value="page" :length="pageCount" density="comfortable" @update:model-value="
                        (p) => {
                            page = p;
                            load();
                        }
                    " />
                </v-card-text>
            </template>
        </v-card>
    </v-container>
</template>

<style scoped>
.filter-bar {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
}

.filter-btn {
    letter-spacing: normal;
    font-weight: 500;
}

.filter-btn :deep(.v-btn__append) {
    margin-inline-start: 4px;
    opacity: 0.7;
}

.filter-clear {
    margin-left: auto;
    opacity: 0.75;
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
    gap: 8px;
    padding: 8px;
}

.listing-card {
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

.listing-card__top {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
}

.listing-card__price {
    font-size: 1.25rem;
    font-weight: 600;
    line-height: 1.2;
}

.listing-card__new {
    margin-left: auto;
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
    margin-top: 10px;
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
    color: rgba(var(--v-theme-on-surface), 0.55);
}
</style>
