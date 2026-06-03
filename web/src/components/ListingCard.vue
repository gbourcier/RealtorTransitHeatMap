<script setup lang="ts">
import { computed, inject } from "vue";
import type { Listing } from "../api/listings";
import { favoritesKey } from "../composables/useFavorites";
import {
    formatPrice,
    formatDate,
    isNew,
    formatCommute,
    commuteColor,
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
            <v-icon size="20">{{ isFavorite ? "mdi-heart" : "mdi-heart-outline" }}</v-icon>
        </button>
        <div class="listing-card__top">
            <span class="listing-card__price">{{ formatPrice(item.currentPrice) }}</span>
            <v-chip
                v-if="!item.isAvailable"
                size="x-small"
                color="warning"
                variant="tonal"
                class="listing-card__expired"
            >expired</v-chip>
            <v-chip
                v-if="isNew(item.firstSeenAt)"
                size="x-small"
                color="secondary"
                :variant="variant === 'panel' ? 'flat' : 'tonal'"
                class="listing-card__new"
            >new</v-chip>
        </div>
        <div class="listing-card__street">
            {{ parseAddress(item.address).street }}
        </div>
        <div v-if="parseAddress(item.address).locality" class="listing-card__locality">
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
                <span
                    class="listing-card__commute-dot"
                    :style="{ background: commuteColor(item.commuteSecondsDowntown) }"
                />
                <span>{{ formatCommute(item.commuteSecondsDowntown) }} downtown</span>
            </a>
            <span v-else class="listing-card__commute listing-card__commute--muted">
                <span
                    class="listing-card__commute-dot"
                    :style="{ background: commuteColor(null) }"
                />
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
    background-color: rgba(var(--v-theme-on-surface), 0.03);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    border-radius: 14px;
    padding: 14px 14px 12px;
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease;
}

.listing-card:active {
    background-color: rgba(var(--v-theme-on-surface), 0.06);
}

.listing-card:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 2px;
}

.listing-card--interactive:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.06);
    border-color: rgba(var(--v-theme-on-surface), 0.18);
}

.listing-card--selected,
.listing-card--selected:hover {
    background-color: rgba(var(--v-theme-primary), 0.12);
    border-color: rgba(var(--v-theme-primary), 0.55);
    box-shadow: inset 0 0 0 1px rgba(var(--v-theme-primary), 0.45);
}

.listing-card--mobile {
    padding: 18px 18px 16px;
}

.listing-card--mobile .listing-card__price {
    font-size: 1.6rem;
    font-weight: 700;
}

.listing-card--mobile .listing-card__street {
    font-size: 1rem;
    font-weight: 500;
}

.listing-card--mobile .listing-card__meta {
    border-top: 0;
    padding-top: 14px;
    margin-top: 10px;
}

.listing-card--mobile .listing-card__new {
    top: 18px;
    right: 58px;
}

.listing-card--mobile .listing-card__fav {
    top: 12px;
    right: 12px;
}

.listing-card__fav {
    position: absolute;
    top: 8px;
    right: 8px;
    z-index: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    border: 0;
    border-radius: 999px;
    background: transparent;
    color: rgba(var(--v-theme-on-surface), 0.5);
    cursor: pointer;
    transition: background-color 120ms ease, color 120ms ease, transform 120ms ease;
}

.listing-card__fav:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.08);
    color: rgba(var(--v-theme-on-surface), 0.85);
}

.listing-card__fav:active {
    transform: scale(0.9);
}

.listing-card__fav:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 2px;
}

.listing-card__fav--active {
    color: rgb(var(--v-theme-accent));
}

.listing-card__fav--active:hover {
    color: rgb(var(--v-theme-accent));
    background-color: rgba(var(--v-theme-accent), 0.12);
}

.listing-card__top {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
    padding-right: 88px;
}

.listing-card__price {
    font-size: 1.25rem;
    font-weight: 600;
    line-height: 1.2;
}

.listing-card__new {
    position: absolute;
    top: 12px;
    right: 50px;
    margin-left: 0;
}

.listing-card__expired {
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
    gap: 8px;
    color: rgba(var(--v-theme-on-surface), 0.85);
    font-weight: 500;
    font-size: 0.8125rem;
}

.listing-card__commute--link {
    text-decoration: none;
    padding: 6px 12px;
    margin: -6px 0 -6px -4px;
    border-radius: 999px;
    background-color: rgba(var(--v-theme-on-surface), 0.06);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    transition: background-color 120ms ease, border-color 120ms ease;
}

.listing-card__commute--link:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.1);
    border-color: rgba(var(--v-theme-on-surface), 0.14);
}

.listing-card__commute--link:active {
    background-color: rgba(var(--v-theme-on-surface), 0.14);
}

.listing-card__commute--link:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 2px;
}

.listing-card__commute--muted {
    color: rgba(var(--v-theme-on-surface), 0.5);
    font-weight: 400;
}

.listing-card__commute-dot {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
    box-shadow: 0 0 0 1px rgba(var(--v-theme-shadow), 0.3);
}

.listing-card__seen {
    margin-left: auto;
    color: rgba(var(--v-theme-on-surface), 0.45);
    font-size: 0.75rem;
}
</style>
