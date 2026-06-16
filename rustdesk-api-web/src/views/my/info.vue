<template>
  <div class="info-container">
    <el-card :title="T('Userinfo')" shadow="hover" class="info-card">
      <template #header>
        <div class="card-header">
          <span>{{ T('Userinfo') }}</span>
        </div>
      </template>
      <el-form class="info-form" ref="form" label-width="150px" label-suffix="：">
        <el-form-item :label="T('Username')">
          <div class="info-value">{{ userStore.username }}</div>
        </el-form-item>
        <el-form-item :label="T('Email')">
          <div class="info-value">{{ userStore.email }}</div>
        </el-form-item>
        <el-form-item :label="T('Password')" prop="password">
          <el-button type="danger" @click="showChangePwd">{{ T('ChangePassword') }}</el-button>
        </el-form-item>
        <el-form-item label="OIDC">
          <el-table :data="oidcData" border fit class="oidc-table">
            <el-table-column :label="T('IdP')" prop="op" align="center"></el-table-column>
            <el-table-column :label="T('Status')" prop="status" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.status === 1" type="success">{{ T('HasBind') }}</el-tag>
                <el-tag v-else type="danger">{{ T('NoBind') }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="T('Actions')" align="center" width="200">
              <template #default="{ row }">
                <el-button v-if="row.status === 1" type="danger" size="small" @click="toUnBind(row)">{{ T('UnBind') }}</el-button>
                <el-button v-else type="success" size="small" @click="toBind(row)">{{ T('ToBind') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="hover" class="config-card">
      <template #header>
        <div class="card-header">
          <span>{{ T('RustDeskConfig') || 'Cấu hình RustDesk Client' }}</span>
        </div>
      </template>
      <el-form class="config-form" label-width="150px" label-suffix="：">
        <el-form-item label="ID Server">
          <el-input :value="appStore.setting.rustdeskConfig.id_server" readonly class="copy-input">
            <template #append>
              <el-button @click="copyText(appStore.setting.rustdeskConfig.id_server)">Copy</el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="Relay Server">
          <el-input :value="appStore.setting.rustdeskConfig.relay_server" readonly class="copy-input">
            <template #append>
              <el-button @click="copyText(appStore.setting.rustdeskConfig.relay_server)">Copy</el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="API Server">
          <el-input :value="appStore.setting.rustdeskConfig.api_server" readonly class="copy-input">
            <template #append>
              <el-button @click="copyText(appStore.setting.rustdeskConfig.api_server)">Copy</el-button>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item :label="T('PublicKey') || 'Khóa công khai (Key)'">
          <el-input :value="appStore.setting.rustdeskConfig.key" type="textarea" :rows="3" readonly class="key-textarea"></el-input>
          <div class="key-actions">
            <el-button type="primary" size="small" @click="copyText(appStore.setting.rustdeskConfig.key)">
              {{ T('CopyKey') || 'Sao chép Khóa công khai' }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="hover" class="hello-card" v-if="appStore.setting.hello">
      <div v-html="html"></div>
    </el-card>

    <changePwdDialog v-model:visible="changePwdVisible"></changePwdDialog>
  </div>
</template>

<script setup>
  import changePwdDialog from '@/components/changePwdDialog.vue'
  import { computed, ref, onMounted } from 'vue'
  import { useUserStore } from '@/store/user'
  import { useAppStore } from '@/store/app'
  import { bind, unbind } from '@/api/oauth'
  import { myOauth } from '@/api/user'
  import { ElMessageBox, ElMessage } from 'element-plus'
  import { T } from '@/utils/i18n'
  import { marked } from 'marked'

  const appStore = useAppStore()
  const userStore = useUserStore()
  const changePwdVisible = ref(false)

  const showChangePwd = () => {
    changePwdVisible.value = true
  }

  const oidcData = ref([])

  const getMyOauth = async () => {
    const res = await myOauth().catch(_ => false)
    if (res) {
      oidcData.value = res.data
    }
  }

  onMounted(() => {
    appStore.loadRustdeskConfig()
    getMyOauth()
  })

  const toBind = async (row) => {
    const res = await bind({ op: row.op }).catch(_ => false)
    if (res) {
      const { code, url } = res.data
      window.open(url)
    }
  }

  const toUnBind = async (row) => {
    const cf = await ElMessageBox.confirm(T('Confirm?', { param: T('UnBind') }), {
      confirmButtonText: T('Confirm'),
      cancelButtonText: T('Cancel'),
      type: 'warning',
    }).catch(_ => false)
    if (!cf) {
      return false
    }
    const res = await unbind({ op: row.op }).catch(_ => false)
    if (res) {
      getMyOauth()
    }
  }

  const copyText = (text) => {
    if (!text) return
    navigator.clipboard.writeText(text).then(() => {
      ElMessage.success(T('CopySuccess') || 'Sao chép thành công!')
    }).catch(() => {
      ElMessage.error(T('CopyFailed') || 'Sao chép thất bại!')
    })
  }

  const html = computed(_ => marked(appStore.setting.hello||''))
</script>

<style scoped lang="scss">
.info-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  font-weight: 600;
  font-size: 16px;
  color: var(--header-text-color);
}

.info-form, .config-form {
  padding: 10px 0;
}

.info-value {
  font-weight: 500;
  color: var(--header-text-color);
}

.oidc-table {
  margin-top: 10px;
}

.copy-input {
  max-width: 450px;
}

.key-textarea {
  max-width: 450px;
}

.key-actions {
  margin-top: 10px;
}

.hello-card {
  line-height: 1.6;
}
</style>

