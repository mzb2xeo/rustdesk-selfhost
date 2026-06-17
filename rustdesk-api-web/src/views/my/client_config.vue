<template>
  <div class="client-config-container">
    <!-- Server Config Card -->
    <el-card shadow="hover" class="config-card">
      <template #header>
        <div class="card-header">
          <el-icon class="header-icon"><Cpu /></el-icon>
          <span>{{ T('RustDeskConfig') || 'Cấu hình RustDesk Client' }}</span>
        </div>
      </template>

      <div class="config-intro">
        <p>{{ T('ConfigIntro') || 'Để thiết bị của bạn có thể kết nối thông qua máy chủ riêng này, vui lòng cấu hình Client RustDesk bằng các thông tin bên dưới.' }}</p>
      </div>

      <el-form class="config-form" label-width="150px" label-position="left">
        <el-form-item label="ID Server">
          <el-input :value="appStore.setting.rustdeskConfig.id_server" readonly class="copy-input">
            <template #append>
              <el-button @click="copyText(appStore.setting.rustdeskConfig.id_server)" type="primary">
                {{ T('Copy') || 'Copy' }}
              </el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="Relay Server">
          <el-input :value="appStore.setting.rustdeskConfig.relay_server" readonly class="copy-input">
            <template #append>
              <el-button @click="copyText(appStore.setting.rustdeskConfig.relay_server)" type="primary">
                {{ T('Copy') || 'Copy' }}
              </el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="API Server">
          <el-input :value="appStore.setting.rustdeskConfig.api_server" readonly class="copy-input">
            <template #append>
              <el-button @click="copyText(appStore.setting.rustdeskConfig.api_server)" type="primary">
                {{ T('Copy') || 'Copy' }}
              </el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item :label="T('PublicKey') || 'Key (Public Key)'">
          <el-input :value="appStore.setting.rustdeskConfig.key" type="textarea" :rows="3" readonly class="key-textarea"></el-input>
          <div class="key-actions">
            <el-button type="primary" size="small" @click="copyText(appStore.setting.rustdeskConfig.key)">
              {{ T('CopyKey') || 'Sao chép Public Key' }}
            </el-button>
          </div>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Auto Deployment Card -->
    <el-card shadow="hover" class="deploy-card">
      <template #header>
        <div class="card-header">
          <el-icon class="header-icon"><Tools /></el-icon>
          <span>{{ T('AutoDeploy') || 'Tự động Cài đặt & Cấu hình (Windows)' }}</span>
        </div>
      </template>

      <div class="deploy-intro">
        <p>{{ T('DeployIntro') || 'Bấm "Tạo lệnh triển khai" để sinh token 1 lần (30 phút). Chạy lệnh PowerShell dưới quyền Administrator trên máy Windows — script sẽ tự tải RustDesk, cấu hình và đăng ký thiết bị vào tài khoản của bạn.' }}</p>
      </div>

      <el-form class="deploy-form" label-width="180px" label-position="left">
        <el-form-item :label="T('DeployPasswordMode') || 'Mat khau remote'">
          <el-radio-group v-model="passwordMode">
            <el-radio label="structured">{{ T('DeployPasswordStructured') || 'Theo cau truc Rd@ + 5 so cuoi ID' }}</el-radio>
            <el-radio label="custom">{{ T('DeployPasswordCustom') || 'Tu nhap mat khau' }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="passwordMode === 'custom'" :label="T('Password') || 'Mat khau'">
          <el-input
            v-model="customPassword"
            type="password"
            show-password
            maxlength="32"
            :placeholder="T('DeployCustomPasswordHint') || '4-32 ky tu'"
            class="copy-input"
          />
        </el-form-item>

        <el-form-item :label="T('DownloadAndRunDeploy') || 'Tu tai script va chay deploy'">
          <el-input :value="downloadRunCommand" type="textarea" :rows="4" readonly class="command-textarea" placeholder="Bấm 'Tạo lệnh triển khai' để sinh lệnh tự tải script và chạy deploy..."></el-input>
          <div class="command-actions">
            <el-button type="primary" :loading="generating" @click="generateDeployCommand">
              {{ T('GenerateDeployCommand') || 'Tạo lệnh triển khai' }}
            </el-button>
            <el-button type="primary" size="small" :disabled="!downloadRunCommand" @click="copyText(downloadRunCommand)">
              {{ T('CopyDownloadRunCommand') || 'Sao chép lệnh tự chạy' }}
            </el-button>
            <el-button size="small" :disabled="!scriptUrl" @click="downloadScript">
              {{ T('DownloadDeployScript') || 'Tải script' }}
            </el-button>
          </div>
        </el-form-item>

        <el-form-item :label="T('DeployCommand') || 'Lệnh PowerShell'">
          <el-input :value="powershellCommand" type="textarea" :rows="4" readonly class="command-textarea" placeholder="Bấm 'Tạo lệnh triển khai' để sinh lệnh mới..."></el-input>
          <div class="command-actions">
            <el-button type="primary" size="small" :disabled="!powershellCommand" @click="copyText(powershellCommand)">
              {{ T('CopyCommand') || 'Sao chép Lệnh' }}
            </el-button>
          </div>
          <p v-if="tokenExpiresAt" class="token-meta">
            Token hết hạn: {{ formatExpire(tokenExpiresAt) }}
          </p>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Guide / Steps Card -->
    <el-card shadow="hover" class="guide-card">
      <template #header>
        <div class="card-header">
          <el-icon class="header-icon"><Notebook /></el-icon>
          <span>{{ T('SetupGuide') || 'Hướng dẫn Cấu hình Nhanh' }}</span>
        </div>
      </template>

      <div class="guide-steps">
        <div class="step-item">
          <div class="step-num">1</div>
          <div class="step-content">
            <h4>{{ T('Step1Title') || 'Tải và cài đặt RustDesk Client' }}</h4>
            <p>{{ T('Step1Desc') || 'Tải xuống ứng dụng RustDesk chính thức từ phần bên dưới phù hợp với hệ điều hành của bạn.' }}</p>
          </div>
        </div>

        <div class="step-item">
          <div class="step-num">2</div>
          <div class="step-content">
            <h4>{{ T('Step2Title') || 'Mở cấu hình Network' }}</h4>
            <p>{{ T('Step2Desc') || 'Mở phần mềm RustDesk -> Click vào biểu tượng Menu (3 dấu gạch/chấm bên cạnh ID của bạn) -> Settings (Thiết lập) -> Network (Mạng).' }}</p>
          </div>
        </div>

        <div class="step-item">
          <div class="step-num">3</div>
          <div class="step-content">
            <h4>{{ T('Step3Title') || 'Điền các thông tin Server & Key' }}</h4>
            <p>{{ T('Step3Desc') || 'Tích chọn "Unlock Network Settings", sau đó sao chép và dán lần lượt ID Server, Relay Server, API Server và Key vào các ô tương ứng rồi bấm Apply (Áp dụng).' }}</p>
          </div>
        </div>
      </div>
    </el-card>

    <!-- Downloads Card -->
    <el-card shadow="hover" class="downloads-card">
      <template #header>
        <div class="card-header">
          <el-icon class="header-icon"><Download /></el-icon>
          <span>{{ T('DownloadClient') || 'Tải Client chính thức' }}</span>
        </div>
      </template>

      <div class="download-grid">
        <a href="https://github.com/rustdesk/rustdesk/releases/latest" target="_blank" class="download-item">
          <div class="os-icon win"></div>
          <div class="download-info">
            <h4>Windows</h4>
            <p>RustDesk official release (.exe / .msi)</p>
          </div>
        </a>

        <a href="https://github.com/rustdesk/rustdesk/releases/latest" target="_blank" class="download-item">
          <div class="os-icon mac"></div>
          <div class="download-info">
            <h4>macOS</h4>
            <p>RustDesk official release (.dmg)</p>
          </div>
        </a>

        <a href="https://github.com/rustdesk/rustdesk/releases/latest" target="_blank" class="download-item">
          <div class="os-icon linux"></div>
          <div class="download-info">
            <h4>Linux</h4>
            <p>Debian, Ubuntu, RedHat (.deb / .rpm)</p>
          </div>
        </a>

        <a href="https://play.google.com/store/apps/details?id=com.carriez.rustdesk" target="_blank" class="download-item">
          <div class="os-icon android"></div>
          <div class="download-info">
            <h4>Android</h4>
            <p>Google Play / Direct APK download</p>
          </div>
        </a>

        <a href="https://apps.apple.com/us/app/rustdesk/id1617462000" target="_blank" class="download-item">
          <div class="os-icon ios"></div>
          <div class="download-info">
            <h4>iOS / iPadOS</h4>
            <p>Official App Store download</p>
          </div>
        </a>
      </div>
    </el-card>
  </div>
</template>

<script setup>
  import { onMounted, ref } from 'vue'
  import { useAppStore } from '@/store/app'
  import { ElMessage } from 'element-plus'
  import { T } from '@/utils/i18n'
  import { Cpu, Notebook, Download, Tools } from '@element-plus/icons'
  import { createDeployToken } from '@/api/my/deploy'

  const appStore = useAppStore()
  const powershellCommand = ref('')
  const downloadRunCommand = ref('')
  const scriptUrl = ref('')
  const tokenExpiresAt = ref(0)
  const generating = ref(false)
  const passwordMode = ref('structured')
  const customPassword = ref('')

  onMounted(() => {
    appStore.loadRustdeskConfig()
  })

  const generateDeployCommand = async () => {
    if (passwordMode.value === 'custom') {
      const pwd = customPassword.value.trim()
      if (pwd.length < 4 || pwd.length > 32) {
        ElMessage.warning(T('DeployCustomPasswordHint') || 'Mat khau tuy chinh phai tu 4-32 ky tu.')
        return
      }
    }
    generating.value = true
    try {
      const res = await createDeployToken({
        password_mode: passwordMode.value,
        custom_password: passwordMode.value === 'custom' ? customPassword.value.trim() : '',
      })
      powershellCommand.value = res.data.powershell_command
      downloadRunCommand.value = res.data.download_run_command || res.data.powershell_command
      scriptUrl.value = res.data.script_url
      tokenExpiresAt.value = res.data.expires_at
      ElMessage.success(T('GenerateDeployCommandSuccess') || 'Đã tạo lệnh triển khai mới!')
    } catch (e) {
      ElMessage.error(T('GenerateDeployCommandFailed') || 'Không thể tạo lệnh triển khai.')
    } finally {
      generating.value = false
    }
  }

  const formatExpire = (ts) => {
    if (!ts) return ''
    return new Date(ts * 1000).toLocaleString()
  }

  const downloadScript = () => {
    if (!scriptUrl.value) return
    window.open(scriptUrl.value, '_blank')
  }

  const copyText = (text) => {
    if (!text) return
    navigator.clipboard.writeText(text).then(() => {
      ElMessage.success(T('CopySuccess') || 'Sao chép thành công!')
    }).catch(() => {
      ElMessage.error(T('CopyFailed') || 'Sao chép thất bại!')
    })
  }
</script>

<style scoped lang="scss">
.client-config-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px 0;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  font-size: 16px;
  color: var(--header-text-color);

  .header-icon {
    font-size: 20px;
    color: #409eff;
  }
}

.config-intro {
  margin-bottom: 20px;
  color: #606266;
  font-size: 14px;
  line-height: 1.6;
}

.config-form {
  padding: 10px 0;
}

.copy-input {
  max-width: 500px;
}

.key-textarea {
  max-width: 500px;
}

.key-actions {
  margin-top: 10px;
}

/* Guide Styles */
.guide-steps {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 10px 0;
}

.step-item {
  display: flex;
  gap: 15px;
  align-items: flex-start;

  .step-num {
    background: #409eff;
    color: #fff;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    flex-shrink: 0;
  }

  .step-content {
    h4 {
      margin: 0 0 5px 0;
      font-size: 15px;
      color: var(--header-text-color);
    }
    p {
      margin: 0;
      font-size: 13.5px;
      color: #606266;
      line-height: 1.5;
    }
  }
}

/* Download Grid Styles */
.download-grid {
  display: grid;
  grid-template-cols: 1fr;
  gap: 15px;
  padding: 10px 0;

  @media (min-width: 768px) {
    grid-template-columns: repeat(2, 1fr);
  }
  @media (min-width: 992px) {
    grid-template-columns: repeat(3, 1fr);
  }
}

.download-item {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  text-decoration: none;
  transition: all 0.3s;
  background-color: var(--el-bg-color);

  &:hover {
    border-color: #409eff;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
    transform: translateY(-2px);
  }

  .os-icon {
    width: 32px;
    height: 32px;
    background-size: contain;
    background-repeat: no-repeat;
    background-position: center;
    flex-shrink: 0;

    &.win {
      background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath fill='%230078d7' d='M0 3.449L9.75 2.1v9.45H0V3.449zM0 12.45h9.75v9.45L0 20.551v-8.1zM10.8 1.95L24 0v11.55H10.8V1.95zM10.8 12.45H24v11.55l-13.2-1.95v-9.6z'/%3E%3C/svg%3E");
    }
    &.mac {
      background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath fill='%238e8e93' d='M17.05 20.28c-.98.95-2.05.88-3.08.4c-1.09-.5-2.08-.48-3.24 0c-1.44.62-2.2.44-3.06-.4C3.8 16.32 3.4 9.64 6.8 9.27c1.3.14 2.16.8 2.84.84c.83.05 1.96-.86 3.53-.74c1.66.12 2.8 1.25 3.32 2.4c-3.1 1.7-2.3 5.7.5 6.9c-.7 1.6-1.5 3.2-2.9 4.8l.04.03l-.04-.02zM12.03 7.25c.15-2 1.8-3.9 3.5-3.8c.2 2-1.5 4-3.5 3.8z'/%3E%3C/svg%3E");
    }
    &.linux {
      background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath fill='%23fca326' d='M12 .007c-3.23 0-5.74 2.65-5.74 5.91a8 8 0 0 0 .53 2.76A7.05 7.05 0 0 0 2.2 14.99A7.06 7.06 0 0 0 9.25 22h5.5A7.06 7.06 0 0 0 21.8 14.99a7.05 7.05 0 0 0-4.59-6.31a8 8 0 0 0 .53-2.76c0-3.26-2.51-5.91-5.74-5.91zm0 2.06c2.04 0 3.68 1.7 3.68 3.85c0 .36-.05.7-.15 1.03a7.22 7.22 0 0 0-7.06 0c-.1-.33-.15-.67-.15-1.03c0-2.15 1.64-3.85 3.68-3.85zm.02 7.7a5 5 0 0 1 5 5a5 5 0 0 1-5 5a5 5 0 0 1-5-5a5 5 0 0 1 5-5zm0 2c-1.66 0-3 1.34-3 3s1.34 3 3 3s3-1.34 3-3s-1.34-3-3-3z'/%3E%3C/svg%3E");
    }
    &.android {
      background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath fill='%23a4c639' d='M5 8v10c0 1.1.9 2 2 2h10c1.1 0 2-.9 2-2V8H5zm7-4.5C9 3.5 6 6.5 6 10h12c0-3.5-3-6.5-6-6.5zm-3.5 11c-.8 0-1.5-.7-1.5-1.5s.7-1.5 1.5-1.5s1.5.7 1.5 1.5s-.7 1.5-1.5 1.5zm7 0c-.8 0-1.5-.7-1.5-1.5s.7-1.5 1.5-1.5s1.5.7 1.5 1.5s-.7 1.5-1.5 1.5zm.9-10.7l.8-.8c.2-.2.2-.5 0-.7a.5.5 0 0 0-.7 0l-.9.9c-.6-.3-1.3-.4-2-.4s-1.4.1-2 .4l-.9-.9a.5.5 0 0 0-.7 0c-.2.2-.2.5 0 .7l.8.8C6.6 6.4 6 7.6 6 9h12c0-1.4-.6-2.6-1.6-3.7z'/%3E%3C/svg%3E");
    }
    &.ios {
      background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath fill='%23007aff' d='M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M15.97 4.17c.66-.81 1.11-1.93.99-3.06-1 .04-2.19.67-2.91 1.49-.62.71-1.16 1.85-1.02 2.96 1.11.09 2.24-.55 2.94-1.39z'/%3E%3C/svg%3E");
    }
  }

  .download-info {
    h4 {
      margin: 0 0 3px 0;
      font-size: 14px;
      color: var(--header-text-color);
    }
    p {
      margin: 0;
      font-size: 11.5px;
      color: #909399;
    }
  }
}

.deploy-card {
  margin-bottom: 0px;
}

.deploy-intro {
  margin-bottom: 20px;
  color: #606266;
  font-size: 14px;
  line-height: 1.6;
}

.deploy-form {
  padding: 10px 0;
}

.command-textarea {
  max-width: 100%;
}

.command-actions {
  margin-top: 10px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.token-meta {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
