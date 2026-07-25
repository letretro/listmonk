<template>
  <div>
    <div class="items zeptomail-settings">
      <div class="block box" v-for="(item, n) in form.zeptomail" :key="n">
        <div class="columns">
          <div class="column is-2">
            <b-field>
              <b-switch v-model="item.enabled" name="enabled" :native-value="true">
                {{ $t('globals.buttons.enabled') }}
              </b-switch>
            </b-field>
            <b-field v-if="form.zeptomail.length > 1">
              <a @click.prevent="$utils.confirm(null, () => removeZeptoMail(n))" href="#">
                <b-icon icon="trash-can-outline" />
                {{ $t('globals.buttons.delete') }}
              </a>
            </b-field>
          </div>

          <div class="column" :class="{ disabled: !item.enabled }">
            <div class="columns">
              <div class="column is-6">
                <b-field :label="$t('globals.fields.name')" label-position="on-border"
                  :message="$t('settings.zeptomail.nameHelp')">
                  <b-input v-model="item.name" name="name" placeholder="zeptomail" :maxlength="100" />
                </b-field>
              </div>
              <div class="column is-6">
                <b-field :label="$t('settings.zeptomail.apiKey')" label-position="on-border"
                  :message="$t('settings.zeptomail.apiKeyHelp')">
                  <b-input v-model="item.api_key" name="api_key" type="password"
                    :placeholder="$t('settings.zeptomail.apiKeyPlaceholder')" :maxlength="500" />
                </b-field>
              </div>
            </div>

            <div class="columns">
              <div class="column is-6">
                <b-field :label="$t('settings.zeptomail.fromEmail')" label-position="on-border"
                  :message="$t('settings.zeptomail.fromEmailHelp')">
                  <b-input v-model="item.from_email" name="from_email" type="email" placeholder="mail@example.com"
                    :maxlength="200" />
                </b-field>
              </div>
              <div class="column is-6">
                <b-field :label="$t('settings.zeptomail.fromName')" label-position="on-border"
                  :message="$t('settings.zeptomail.fromNameHelp')">
                  <b-input v-model="item.from_name" name="from_name" placeholder="Sender Name" :maxlength="200" />
                </b-field>
              </div>
            </div>

            <hr />

            <div class="columns">
              <div class="column">
                <b-field>
                  <b-switch v-model="item.track_opens" name="track_opens">
                    {{ $t('settings.zeptomail.trackOpens') }}
                  </b-switch>
                </b-field>
              </div>
              <div class="column">
                <b-field>
                  <b-switch v-model="item.track_clicks" name="track_clicks">
                    {{ $t('settings.zeptomail.trackClicks') }}
                  </b-switch>
                </b-field>
              </div>
            </div>

            <div class="columns">
              <div class="column is-4">
                <b-field :label="$t('settings.zeptomail.timeout')" label-position="on-border"
                  :message="$t('settings.zeptomail.timeoutHelp')">
                  <b-input v-model="item.timeout" name="timeout" placeholder="30s" :pattern="regDuration"
                    :maxlength="10" />
                </b-field>
              </div>
              <div class="column is-4">
                <b-field :label="$t('settings.zeptomail.retries')" label-position="on-border"
                  :message="$t('settings.zeptomail.retriesHelp')">
                  <b-numberinput v-model="item.max_msg_retries" name="max_msg_retries" type="is-light"
                    controls-position="compact" placeholder="2" min="0" max="100" />
                </b-field>
              </div>
            </div>

            <hr />

            <form @submit.prevent="() => doZeptoMailTest(item, n)">
              <div class="columns">
                <template v-if="testItem === n">
                  <div class="column is-5">
                    <strong>{{ $t('settings.zeptomail.fromEmail') }}</strong>
                    <br />
                    {{ item.from_email || settings['app.from_email'] }}
                  </div>
                  <div class="column is-4">
                    <b-field :label="$t('settings.zeptomail.toEmail')" label-position="on-border">
                      <b-input type="email" required v-model="testEmail" :ref="'testEmailTo'"
                        placeholder="email@site.com" :custom-class="`zeptomail-test-email-${n}`" />
                    </b-field>
                  </div>
                </template>
                <div class="column has-text-right">
                  <b-button v-if="testItem === n" class="is-primary" @click.prevent="() => doZeptoMailTest(item, n)">
                    {{ $t('settings.zeptomail.sendTest') }}
                  </b-button>
                  <a href="#" v-else class="is-primary" @click.prevent="showTestForm(n)">
                    <b-icon icon="rocket-launch-outline" /> {{ $t('settings.zeptomail.testConnection') }}
                  </a>
                </div>
              </div>
              <div v-if="errMsg && testItem === n">
                <b-field class="mt-4" type="is-danger">
                  <b-input v-model="errMsg" type="textarea" custom-class="has-text-danger is-size-6" readonly />
                </b-field>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>

    <b-button @click="addZeptoMail" icon-left="plus" type="is-primary">
      {{ $t('globals.buttons.addNew') }}
    </b-button>
  </div>
</template>

<script>
import Vue from 'vue';
import { mapState } from 'vuex';
import { regDuration } from '../../constants';

export default Vue.extend({
  props: {
    form: {
      type: Object, default: () => ({}),
    },
  },

  data() {
    return {
      data: this.form,
      regDuration,
      testItem: null,
      testEmail: '',
      errMsg: '',
    };
  },

  methods: {
    addZeptoMail() {
      this.data.zeptomail.push({
        name: '',
        enabled: true,
        api_key: '',
        from_email: '',
        from_name: '',
        track_opens: true,
        track_clicks: true,
        timeout: '30s',
        max_msg_retries: 2,
      });

      this.$nextTick(() => {
        const items = document.querySelectorAll('.zeptomail-settings input[name="name"]');
        items[items.length - 1].focus();
      });
    },

    removeZeptoMail(i) {
      this.data.zeptomail.splice(i, 1);
    },

    showTestForm(n) {
      this.testItem = n;
      this.errMsg = '';

      this.$nextTick(() => {
        const el = document.querySelector(`.zeptomail-test-email-${n}`);
        if (el) el.focus();
      });
    },

    doZeptoMailTest(item, n) {
      if (!this.isTestEnabled(item)) {
        this.$utils.toast(this.$t('settings.zeptomail.testEnterKey'), 'is-danger');
        this.$nextTick(() => {
          const inputs = document.querySelectorAll('.zeptomail-settings input[name="api_key"]');
          inputs[n].focus();
          inputs[n].select();
        });
        return;
      }

      this.errMsg = '';
      this.$api.testZeptoMail({ ...item, email: this.testEmail }).then(() => {
        this.$utils.toast(this.$t('campaigns.testSent'));
      }).catch((err) => {
        if (err.response?.data?.message) {
          this.errMsg = err.response.data.message;
        }
      });
    },

    isTestEnabled(item) {
      if (item.api_key.includes('•')) {
        return false;
      }
      return true;
    },
  },

  computed: {
    ...mapState(['settings']),
  },
});
</script>
