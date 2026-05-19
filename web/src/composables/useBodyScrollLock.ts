import { onMounted, onBeforeUnmount } from "vue";

export function useBodyScrollLock(): void {
    let prevHtmlOverflow = "";
    let prevBodyOverflow = "";

    onMounted(() => {
        prevHtmlOverflow = document.documentElement.style.overflow;
        prevBodyOverflow = document.body.style.overflow;
        document.documentElement.style.overflow = "hidden";
        document.body.style.overflow = "hidden";
    });

    onBeforeUnmount(() => {
        document.documentElement.style.overflow = prevHtmlOverflow;
        document.body.style.overflow = prevBodyOverflow;
    });
}
