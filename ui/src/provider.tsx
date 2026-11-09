import React, { useEffect, useState } from "react";
import { Alert, Button, Form, Input, Switch, Typography } from "antd";
import { PLUGIN_ID, sdk } from "./host";

type Flags = {
  enabled?: boolean;
  package_name?: string;
  pubsub_topic?: string;
  skip_remote_verify?: boolean;
  has_service_account?: boolean;
};

export default function ProviderTab() {
  const host = sdk();
  const projectId = host.projectId();
  const [form] = Form.useForm();
  const [flags, setFlags] = useState<Flags>({});
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!projectId) return;
    void host
      .rest(PLUGIN_ID)
      .get("/config")
      .then((body) => setFlags((body || {}) as Flags))
      .catch(() => setFlags({}));
  }, [host, projectId]);

  useEffect(() => {
    form.setFieldsValue({
      enabled: flags.enabled ?? true,
      package_name: flags.package_name ?? "",
      pubsub_topic: flags.pubsub_topic ?? "",
      skip_remote_verify: flags.skip_remote_verify ?? false,
    });
  }, [flags, form]);

  if (!projectId) {
    return <Alert type="warning" showIcon message="Select a project to configure Google Play." />;
  }

  return (
    <div>
      <Typography.Paragraph type="secondary">
        Map each plan with provider_products provider <Typography.Text code>google_play</Typography.Text>.
        Leave the service account blank to keep the saved JSON.
        {flags.has_service_account ? " A service account is saved." : ""}
      </Typography.Paragraph>
      <Form
        form={form}
        layout="vertical"
        onFinish={async (values) => {
          setSaving(true);
          try {
            const env: Record<string, string> = {
              PLAY_ENABLED: values.enabled ? "true" : "false",
              PLAY_PACKAGE_NAME: values.package_name || "",
              PLAY_PUBSUB_TOPIC: values.pubsub_topic || "",
              PLAY_SKIP_REMOTE_VERIFY: values.skip_remote_verify ? "true" : "false",
            };
            if (values.service_account_json) env.PLAY_SERVICE_ACCOUNT_JSON = values.service_account_json;
            await host.saveProjectEnv(PLUGIN_ID, env);
            const body = (await host.rest(PLUGIN_ID).get("/config")) as Flags;
            setFlags(body || {});
            form.setFieldsValue({ service_account_json: "" });
            host.notify.success("Google Play settings saved");
          } catch (err) {
            host.notify.error(err instanceof Error ? err.message : "Save failed");
          } finally {
            setSaving(false);
          }
        }}
      >
        <Form.Item name="enabled" label="Enabled" valuePropName="checked">
          <Switch />
        </Form.Item>
        <Form.Item name="package_name" label="Package name">
          <Input placeholder="com.example.app" />
        </Form.Item>
        <Form.Item name="pubsub_topic" label="Pub/Sub topic">
          <Input />
        </Form.Item>
        <Form.Item name="service_account_json" label="Service account JSON">
          <Input.TextArea rows={4} placeholder={flags.has_service_account ? "Saved" : ""} />
        </Form.Item>
        <Form.Item name="skip_remote_verify" label="Skip Play API (development)" valuePropName="checked">
          <Switch />
        </Form.Item>
        <Button type="primary" htmlType="submit" loading={saving}>
          Save Google Play
        </Button>
      </Form>
    </div>
  );
}
