<template>
  <el-icon class="ex-icon" @click="expandOrFoldSlider">
    <el-icon-expand v-if="setting.sideIsCollapse"></el-icon-expand>
    <el-icon-fold v-else></el-icon-fold>
  </el-icon>
  <div class="header-logo">
    <img :src="setting.logo" alt="" class="logo">
    <div class="title">{{setting.title}}</div>
  </div>
  <Setting></Setting>
</template>

<script>
  import { defineComponent, computed } from 'vue'
  import HeaderMenu from '@/layout/components/menu/index.vue'
  import Setting from '@/layout/components/setting/index.vue'
  import { useAppStore } from '@/store/app'
  import GTags from '@/layout/components/tags/index.vue'

  export default defineComponent({
    name: 'LayerHeader',
    created () {
    },
    components: { HeaderMenu, Setting, GTags },
    watch: {},
    setup (props) {
      const appStore = useAppStore()
      const setting = computed(() => appStore.setting)
      const expandOrFoldSlider = () => {
        appStore.sideCollapse()
      }
      return {
        setting,
        expandOrFoldSlider,
      }
    },

  })
</script>

<style scoped lang="scss">
  .ex-icon {
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: auto 12px auto 14px;
    font-size: 18px;
    cursor: pointer;
    border-radius: 8px;
    color: var(--app-text-muted);
    transition: background-color 0.18s ease, color 0.18s ease, transform 0.18s ease;

    &:hover {
      color: var(--primaryColor);
      background: var(--el-menu-hover-bg-color);
    }

    &:active {
      transform: translateY(1px);
    }
  }

  .header-logo {
    display: flex;
    height: 100%;
    align-items: center;

    .title {
      display: block;
      margin-left: 12px;
      font-weight: 700;
      letter-spacing: 0;
      color: var(--header-text-color);
      white-space: nowrap;
    }

    .logo {
      display: block;
      width: 34px;
      height: 34px;
      border-radius: 9px;
      box-shadow: 0 8px 18px rgba(37, 99, 235, 0.18);
    }
  }


</style>
<style lang="scss">

</style>
