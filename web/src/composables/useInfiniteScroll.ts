import { watch, onBeforeUnmount, type Ref } from "vue";

export function useInfiniteScroll(
    sentinel: Ref<HTMLElement | null>,
    root: Ref<HTMLElement | null> | null,
    onIntersect: () => void,
): void {
    let observer: IntersectionObserver | null = null;

    function rebuild(sentinelEl: HTMLElement | null, rootEl: HTMLElement | null) {
        observer?.disconnect();
        observer = null;
        if (!sentinelEl) return;
        observer = new IntersectionObserver(
            (entries) => {
                if (entries.some((e) => e.isIntersecting)) onIntersect();
            },
            { root: rootEl, rootMargin: "200px" },
        );
        observer.observe(sentinelEl);
    }

    if (root) {
        watch(
            [sentinel, root],
            ([s, r]) => rebuild(s, r),
            { flush: "post" },
        );
    } else {
        watch(
            sentinel,
            (s) => rebuild(s, null),
            { flush: "post" },
        );
    }

    onBeforeUnmount(() => observer?.disconnect());
}
