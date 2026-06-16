<template>
  <div class="login-container">
    <div class="login-card">
      <div class="brand">
        <img src="@/assets/logo.png" alt="logo" class="login-logo"/>
        <div>
          <h1>RustDesk API</h1>
          <p>Admin console</p>
        </div>
      </div>

      <el-form v-if="!disablePwd" label-position="top" class="login-form">
        <el-form-item :label="T('Username')">
          <el-input v-model="form.username" type="username" class="login-input"></el-input>
        </el-form-item>

        <el-form-item :label="T('Password')">
          <el-input v-model="form.password" type="password" @keyup.enter.native="login" show-password
                    class="login-input"></el-input>
        </el-form-item>
        <el-form-item :label="T('Captcha')" v-if="captchaCode">
          <el-input v-model="form.captcha" @keyup.enter.native="login"  class="login-input captcha-input">
            <template #append>
              <img :src="captchaCode.b64" @click="loadCaptcha" class="captcha" alt="captcha"/>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item>
          <div class="login-actions">
            <el-button @click="login" type="primary" class="login-button">{{ T('Login') }}</el-button>
            <el-button v-if="allowRegister" @click="register" class="login-button secondary">{{ T('Register') }}</el-button>
          </div>
        </el-form-item>
      </el-form>

      <div class="divider" v-if="options.length > 0 && !disablePwd">
        <span>{{ T('or login in with') }}</span>
      </div>

      <div class="oidc-options">
        <div v-for="(option, index) in options" :key="index" class="oidc-option">
          <el-button @click="handleOIDCLogin(option.name)" class="oidc-btn">
            <img :src="getProviderImage(option.name)" alt="provider" class="oidc-icon"/>
            <span>{{ T(option.name) }}</span>
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { reactive, onMounted, ref } from 'vue'
  import { useUserStore } from '@/store/user'
  import { ElMessage } from 'element-plus'
  import { T } from '@/utils/i18n'
  import { useRoute, useRouter } from 'vue-router'
  import { loginOptions, captcha } from '@/api/login'
  import { getCode, removeCode } from '@/utils/auth'

  const oauthInfo = ref({})
  const userStore = useUserStore()
  const route = useRoute()
  const router = useRouter()
  const options = reactive([]) // Storing OIDC login options

  let platform = window.navigator.platform
  if (navigator.platform.indexOf('Mac') === 0) {
    platform = 'mac'
  } else if (navigator.platform.indexOf('Win') === 0) {
    platform = 'windows'
  } else if (navigator.platform.indexOf('Linux armv') === 0) {
    platform = 'android'
  } else if (navigator.platform.indexOf('Linux') === 0) {
    platform = 'linux'
  }
  const userAgent = navigator.userAgent
  let browser = 'Unknown Browser'
  if (/chrome|crios/i.test(userAgent)) browser = 'Chrome'
  else if (/firefox|fxios/i.test(userAgent)) browser = 'Firefox'
  else if (/safari/i.test(userAgent) && !/chrome/i.test(userAgent)) browser = 'Safari'
  else if (/edg/i.test(userAgent)) browser = 'Edge'

  const form = reactive({
    username: '',
    password: '',
    platform: platform,
    captcha: '',
    captcha_id: ''
  })

  const captchaCode = ref('')
  const redirect = route.query?.redirect
  const login = async () => {
    const res = await userStore.login(form).catch(e => e)
    if (!res.code) {
      ElMessage.success(T('LoginSuccess'))
      router.push({ path: redirect || '/', replace: true })
      return
    }
    if (res.code === 110) {
      // need captcha
      loadCaptcha()
    }
  }

  const loadCaptcha = async () => {
    const captchaRes = await captcha().catch(_ => false)
    console.log(captchaRes)
    captchaCode.value = captchaRes.data.captcha
    form.captcha_id = captchaRes.data.captcha.id
  }

  const handleOIDCLogin = (provider) => {
    userStore.oidc(provider, platform, browser)
  }

  import googleImage from '@/assets/google.png'
  import githubImage from '@/assets/github.png'
  import oidcImage from '@/assets/oidc.png'
  import webauthImage from '@/assets/webauth.png'
  import defaultImage from '@/assets/oidc.png'

  const providerImageMap = {
    google: googleImage,
    github: githubImage,
    oidc: oidcImage,
    // WebAuth: webauthImage,
    default: defaultImage,
  }

  const getProviderImage = (provider) => {
    return providerImageMap[provider.toLowerCase()] || providerImageMap.default
  }

  const allowRegister = ref(false)
  const disablePwd = ref(false)
  const loadLoginOptions = async () => {
    try {
      const res = await loginOptions().catch(_ => false)
      if (!res || !res.data) return console.error('No valid response received')
      res.data.ops.map(option => (options.push({ name: option }))) // Create new object array
      if (res.data.auto_oidc) {
        // If there is an automatic OIDC login option, call the first one directly
        handleOIDCLogin(res.data.ops[0])
      }
      disablePwd.value = res.data.disable_pwd
      allowRegister.value = res.data.register
      if (res.data.need_captcha) {
        loadCaptcha()
      }
    } catch (error) {
      console.error('Error loading login options:', error.message)
    }
  }

  onMounted(async () => {
    const code = getCode()
    if (code) {
      // If the code exists, perform a query to obtain user information.
      const res = await userStore.query(code)
      if (res) {
        // Delete the code and make sure to clear the code before jumping
        removeCode()
        ElMessage.success(T('LoginSuccess'))
        router.push({ path: redirect || '/', replace: true })
      }
    } else {
      // If the code does not exist, the login page will be displayed.
      loadLoginOptions() // After the component is mounted, the login option loading function is called
    }
  })

  const register = () => {
    router.push('/register')
  }
</script>

<style scoped lang="scss">
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100dvh;
  background:
    radial-gradient(circle at 50% 0%, rgba(37, 99, 235, 0.14), transparent 26rem),
    linear-gradient(180deg, #f8fafc 0%, #eef4ff 100%);
  padding: 24px;
  box-sizing: border-box;
  position: relative;
  overflow: hidden;
}

.login-card {
  width: min(100%, 410px);
  background-color: #ffffff;
  padding: 32px;
  border-radius: 12px;
  border: 1px solid #dbe4f0;
  box-shadow: 0 22px 60px rgba(15, 23, 42, 0.12);
  position: relative;
  z-index: 1;
}

h1 {
  margin: 0;
  font-size: 24px;
  line-height: 1.1;
  font-weight: 800;
  color: #0f172a;
}

.brand {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 26px;

  p {
    margin: 5px 0 0;
    color: #64748b;
    font-size: 14px;
    font-weight: 650;
  }
}

.login-form {
  margin-bottom: 16px;
}

.login-input {
  width: 100%;

  .captcha{
    cursor: pointer;
    width: 150px;
  }
}
.captcha-input{
  :deep(.el-input-group__append) {
    border-radius: 0 8px 8px 0;
    padding: 0;
    overflow: hidden;
  }
}

.login-actions {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
  width: 100%;
}

.login-button {
  width: 100%;
  height: 42px;
  margin-left: 0;

  &.secondary {
    margin-left: 0;
  }
}

.divider {
  display: flex;
  align-items: center;
  margin: 20px 0;
  font-size: 14px;
  color: #64748b;
  font-weight: 650;

  &::before,
  &::after {
    content: '';
    flex: 1;
    height: 1px;
    background-color: #e2e8f0;
  }

  &::before {
    margin-right: 10px;
  }

  &::after {
    margin-left: 10px;
  }
}

.oidc-options {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.oidc-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  width: 100%;
  height: 50px;
  background-color: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  color: #0f172a;
  font-size: 14px;
  font-weight: 650;
  margin-left: 0;
}

.oidc-icon {
  width: 24px;
  height: 24px;
  margin-right: 0;
}

.login-logo {
  width: 54px;
  height: 54px;
  margin: 0;
  display: block;
  border-radius: 12px;
  box-shadow: 0 14px 28px rgba(37, 99, 235, 0.20);
}

.el-form-item {
  margin-bottom: 18px;

  &:last-child {
    margin-bottom: 0;
  }

  ::v-deep(.el-form-item__label) {
    color: #334155;
    font-weight: 700;
    line-height: 1.2;
    margin-bottom: 8px;
  }

  .el-input {
    ::v-deep(.el-input__wrapper) {
      min-height: 42px;
      border: 0;
      background: #f8fafc;
    }

    ::v-deep(input) {
      color: #0f172a;
    }
  }
}

@media (max-width: 480px) {
  .login-container {
    padding: 16px;
  }

  .login-card {
    padding: 24px;
  }
}
</style>
