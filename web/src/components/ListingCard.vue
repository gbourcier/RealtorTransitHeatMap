<script setup lang="ts">
import { computed, inject } from "vue";
import type { Listing } from "../api/listings";
import { favoritesKey } from "../composables/useFavorites";
import {
    formatPrice,
    formatPropertyType,
    formatDate,
    isNew,
    formatCommute,
    commuteMapUrl,
    parseAddress,
} from "../utils/listingFormat";

interface Props {
    item: Listing;
    variant: "panel" | "mobile";
    selected?: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
    click: [item: Listing];
    hover: [item: Listing];
    leave: [];
}>();

const favorites = inject(favoritesKey, null);

const isFavorite = computed(() =>
    favorites
        ? favorites.isFavorite(props.item.board, props.item.mls, props.item.isFavorite)
        : props.item.isFavorite,
);

const address = computed(() => parseAddress(props.item.address));
const propertyType = computed(() => formatPropertyType(props.item.buildingType));
const commuteTierClass = computed(() => {
    if (props.item.commuteSecondsDowntown == null) return "listing-card__commute--unknown";
    const minutes = props.item.commuteSecondsDowntown / 60;
    if (minutes < 30) return "listing-card__commute--fast";
    if (minutes <= 60) return "listing-card__commute--mid";
    return "listing-card__commute--slow";
});

function onActivate() {
    emit("click", props.item);
}

function onToggleFavorite() {
    favorites?.toggle({
        board: props.item.board,
        mls: props.item.mls,
        isFavorite: props.item.isFavorite,
    });
}
</script>

<template>
    <div
        :role="variant === 'panel' ? 'button' : 'link'"
        tabindex="0"
        class="listing-card"
        :class="{
            'listing-card--interactive': variant === 'panel',
            'listing-card--selected': variant === 'panel' && selected,
            'listing-card--mobile': variant === 'mobile',
        }"
        @click="onActivate"
        @keydown.enter.prevent="onActivate"
        @keydown.space.prevent="onActivate"
        @mouseenter="variant === 'panel' && emit('hover', item)"
        @mouseleave="variant === 'panel' && emit('leave')"
        @focus="variant === 'panel' && emit('hover', item)"
        @blur="variant === 'panel' && emit('leave')"
    >
        <div class="listing-card__top">
            <div class="listing-card__header">
                <div class="listing-card__price">{{ formatPrice(item.currentPrice) }}</div>
                <div class="listing-card__street">
                    {{ address.street }}
                </div>
                <div v-if="address.locality" class="listing-card__locality">
                    {{ address.locality }}
                </div>
            </div>
            <div v-if="propertyType || isNew(item.firstSeenAt) || !item.isAvailable || favorites" class="listing-card__actions">
                <span
                    v-if="propertyType"
                    class="listing-card__typebadge"
                >{{ propertyType }}</span>
                <span
                    v-if="!item.isAvailable"
                    class="listing-card__typebadge listing-card__typebadge--warning"
                >expired</span>
                <span
                    v-else-if="isNew(item.firstSeenAt)"
                    class="listing-card__typebadge listing-card__typebadge--new"
                >new</span>
                <button
                    v-if="favorites"
                    type="button"
                    class="listing-card__fav"
                    :class="{ 'listing-card__fav--active': isFavorite }"
                    :aria-pressed="isFavorite"
                    :aria-label="isFavorite ? 'Remove from favorites' : 'Add to favorites'"
                    @click.stop="onToggleFavorite"
                    @keydown.enter.stop="onToggleFavorite"
                    @keydown.space.stop.prevent="onToggleFavorite"
                >
                    <v-icon size="19">{{ isFavorite ? "mdi-heart" : "mdi-heart-outline" }}</v-icon>
                </button>
            </div>
        </div>
        <div class="listing-card__meta">
            <a
                v-if="item.commuteSecondsDowntown != null && item.address"
                :href="commuteMapUrl(item.address) ?? '#'"
                target="_blank"
                rel="noopener noreferrer"
                class="listing-card__commute listing-card__commute--link"
                :class="commuteTierClass"
                @click.stop
            >
                <span class="listing-card__commute-dot" />
                <span><b>{{ formatCommute(item.commuteSecondsDowntown) }}</b> downtown</span>
            </a>
            <span v-else class="listing-card__commute listing-card__commute--muted" :class="commuteTierClass">
                <span class="listing-card__commute-dot" />
                —
            </span>
            <span class="listing-card__seen">{{ formatDate(item.firstSeenAt) }}</span>
        </div>
    </div>
</template>

<style scoped>
.listing-card {
    position: relative;
    display: block;
    text-decoration: none;
    color: inherit;
    background-color: transparent;
    border: 0;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-radius: 0;
    padding: 16px;
    cursor: pointer;
    transition: background-color 120ms ease, box-shadow 120ms ease;
}

.listing-card:active {
    background-color: rgba(var(--v-theme-on-surface), 0.05);
}

.listing-card:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 2px;
}

.listing-card--interactive:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.035);
}

.listing-card--selected,
.listing-card--selected:hover {
    background-color: rgba(var(--v-theme-primary), 0.09);
    box-shadow: inset 3px 0 0 rgb(var(--v-theme-primary));
}

.listing-card--mobile {
    padding: 18px;
}

.listing-card--mobile .listing-card__price {
    font-size: 1.6rem;
}

.listing-card--mobile .listing-card__street {
    font-size: 1rem;
    font-weight: 500;
}

.listing-card--mobile .listing-card__meta {
    border-top: 0;
    padding-top: 0;
    margin-top: 14px;
}

.listing-card__top {
    display: flex;
    align-items: flex-start;
    gap: 12px;
}

.listing-card__header {
    flex: 1 1 auto;
    min-width: 0;
}

.listing-card__actions {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: -3px;
}

.listing-card__fav {
    display: grid;
    place-items: center;
    width: 32px;
    height: 32px;
    border: 0;
    border-radius: 999px;
    background: transparent;
    color: rgba(var(--v-theme-on-surface), 0.34);
    cursor: pointer;
    transition: background-color 120ms ease, color 120ms ease;
}

.listing-card__fav:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.08);
    color: rgba(var(--v-theme-on-surface), 0.52);
}

.listing-card__fav:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 2px;
}

.listing-card__fav--active {
    color: rgb(var(--v-theme-error));
}

.listing-card__fav--active:hover {
    color: rgb(var(--v-theme-error));
    background-color: rgba(var(--v-theme-error), 0.08);
}

.listing-card__price {
    margin-bottom: 8px;
    font-size: 23px;
    font-weight: 800;
    letter-spacing: 0;
    font-variant-numeric: tabular-nums;
    line-height: 1;
}

.listing-card__typebadge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 24px;
    padding: 0 10px;
    border-radius: 999px;
    background: rgba(var(--v-theme-on-surface), 0.06);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    font-size: 12px;
    font-weight: 600;
    color: rgba(var(--v-theme-on-surface), 0.52);
    white-space: nowrap;
}

.listing-card__typebadge--new {
    color: rgb(var(--v-theme-secondary));
    border-color: rgba(var(--v-theme-secondary), 0.28);
    background: rgba(var(--v-theme-secondary), 0.08);
}

.listing-card__typebadge--warning {
    color: rgb(var(--v-theme-warning));
    border-color: rgba(var(--v-theme-warning), 0.32);
    background: rgba(var(--v-theme-warning), 0.09);
}

.listing-card__street {
    font-size: 16px;
    font-weight: 700;
    letter-spacing: 0;
    line-height: 1.25;
    overflow-wrap: anywhere;
}

.listing-card__locality {
    margin-top: 3px;
    font-size: 13px;
    line-height: 1.3;
    color: rgba(var(--v-theme-on-surface), 0.52);
    font-weight: 500;
    overflow-wrap: anywhere;
}

.listing-card__meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-top: 14px;
    font-size: 12.5px;
}

.listing-card__commute {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 28px;
    padding: 0 12px 0 10px;
    border-radius: 999px;
    font-size: 12.5px;
    font-weight: 600;
    white-space: nowrap;
    --commute-c: var(--v-theme-on-surface);
    background: color-mix(in srgb, rgb(var(--commute-c)) 13%, transparent);
    border: 1px solid color-mix(in srgb, rgb(var(--commute-c)) 38%, transparent);
    color: color-mix(in srgb, rgb(var(--commute-c)) 62%, rgb(var(--v-theme-on-surface)));
}

.listing-card__commute--link {
    text-decoration: none;
    transition: filter 120ms ease;
}

.listing-card__commute--link:hover {
    filter: brightness(1.08);
}

.listing-card__commute--link:active {
    filter: brightness(1.16);
}

.listing-card__commute--link:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 2px;
}

.listing-card__commute--muted {
    --commute-c: var(--v-theme-on-surface);
    color: rgba(var(--v-theme-on-surface), 0.5);
}

.listing-card__commute-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex: 0 0 auto;
    background: rgb(var(--commute-c));
    box-shadow: 0 0 0 3px color-mix(in srgb, rgb(var(--commute-c)) 20%, transparent);
}

.listing-card__commute b {
    font-weight: 800;
    font-variant-numeric: tabular-nums;
    color: color-mix(in srgb, rgb(var(--commute-c)) 80%, rgb(var(--v-theme-on-surface)));
}

.listing-card__commute--fast {
    --commute-c: var(--v-theme-commute-fast);
}

.listing-card__commute--mid {
    --commute-c: var(--v-theme-commute-mid);
}

.listing-card__commute--slow {
    --commute-c: var(--v-theme-commute-slow);
}

.listing-card__seen {
    margin-left: auto;
    color: rgba(var(--v-theme-on-surface), 0.34);
    font-size: 12px;
    font-weight: 500;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
}
</style>
