<script>
import Button from "primevue/button";
import Card from "primevue/card";
import InputText from "primevue/inputtext";
import ProgressSpinner from "primevue/progressspinner";
import apiService from "@/services/apiService";

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

  methods: {
    async loginClick() {
      const response = await apiService.login(this.username, this.password);

      if (response) {
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
          <Button label="Login" @click="loginClick" />
        </div>
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
