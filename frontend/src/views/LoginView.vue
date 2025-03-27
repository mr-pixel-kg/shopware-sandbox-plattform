<script>
import Button from "primevue/button";
import Card from "primevue/card";
import InputText from "primevue/inputtext";
import ProgressSpinner from "primevue/progressspinner";
import apiService from "@/services/apiService";
import {GeneralStore} from "@/stores/generalStore.js";
import {useAuthStore} from "@/stores/authStore.js";

export default {
  components: {
    Card,
    Button,
    InputText,
    ProgressSpinner,
  },

  data() {
    return {
      username: "",
      password: "",
    };
  },

  setup() {
    const authStore = useAuthStore()
    return {
      authStore: authStore,
    };
  },

  methods: {
    async loginClick() {
      const success = await this.authStore.login(this.username, this.password);

      if (success) {
        this.$router.push("/admin");
      } else {
        this.password = "";
      }
    },
  },
};
</script>

<template>
  <div class="login-form">
    <Card>
      <template #title>Login</template>
      <template #content>
        <form @submit.prevent="loginClick">
          <div class="flex justify-center flex-col gap-4">
            <div class="flex flex-col gap-1">
              <InputText
                name="username"
                type="text"
                placeholder="Username"
                v-model="username"
              />
            </div>
            <div class="flex flex-col gap-1">
              <InputText
                name="password"
                type="password"
                placeholder="Password"
                v-model="password"
              />
            </div>
            <Button label="Login" type="submit" />
          </div>
        </form>
      </template>
    </Card>
  </div>
</template>

<style scoped>
.login-form {
  max-width: 400px;
  margin: 50px auto 0;
}
</style>
