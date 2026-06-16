<template>
  <el-scrollbar class="scroll-sidebar" height="100dvh">
    <menus></menus>
  </el-scrollbar>
</template>
<script>
  import Menus from '@/layout/components/menu/index.vue'
  import { defineComponent, ref, onMounted } from 'vue'

  export default defineComponent({
    name: 'GAside',
    components: { Menus },
  })
</script>

<style scoped lang="scss">
.scroll-sidebar {
  background-color: var(--side-bg-color);
  height: 100dvh;
  box-shadow: 8px 0 28px rgba(15, 23, 42, 0.04);
}
</style>
