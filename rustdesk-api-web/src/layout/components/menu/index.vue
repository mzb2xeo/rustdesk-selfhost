<template>
  <el-menu
          class="menus"
          :collapse="isCollapse"
          :default-active="activeIndex"
          router
  >
    <menu-item v-for="(route,index) in routes" :key="route.name" :route="route"></menu-item>
  </el-menu>
</template>

<script>
  import { defineComponent, ref, onMounted, watch, computed } from 'vue'
  import { useRouteStore } from '@/store/router'
  import MenuItem from '@/layout/components/menu/item.vue'
  import { useRoute } from 'vue-router'
  import { useAppStore } from '@/store/app'

  export default defineComponent({
    name: 'Menu',
    created () {
    },
    components: { MenuItem },
    setup () {
      const routes = ref([])
      const route = useRoute()
      const app = useAppStore()
      const isCollapse = computed(() => app.setting.sideIsCollapse)
      const activeIndex = computed(() => route.name)

      routes.value = useRouteStore().routes
      return {
        routes,
        activeIndex,
        isCollapse,
      }
    },

  })
</script>

<style lang="scss" scoped>
  .menus {
    min-height: 100dvh;
    border-right: none;
    padding: 10px 8px;

    &:not(.el-menu--collapse) {
      width: var(--sideBarWidth);
    }
  }

  :deep(.el-menu-item),
  :deep(.el-sub-menu__title) {
    border-radius: 8px;
    margin: 2px 0;
    font-weight: 600;
  }

  :deep(.el-menu-item.is-active) {
    background: #eff6ff;
    box-shadow: inset 3px 0 0 var(--primaryColor);
  }

  :deep(.el-icon) {
    font-size: 18px;
  }
</style>
<style>
</style>
