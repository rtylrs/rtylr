//! S3 object storage, mirroring `internal/platform/storage/r2.go`.

use thiserror::Error;

pub const DEFAULT_FLOCI_ENDPOINT: &str = "http://localhost:4566";
pub const DEFAULT_FLOCI_BUCKET: &str = "onion-dev";
pub const DEFAULT_FLOCI_REGION: &str = "us-east-1";

#[derive(Debug, Error, PartialEq)]
pub enum StorageError {
    #[error("s3 storage is not configured")]
    NotConfigured,
    #[error("s3: {0}")]
    S3(String),
}

#[derive(Debug, Clone, PartialEq)]
pub struct Settings {
    pub endpoint: String,
    pub region: String,
    pub bucket: String,
    pub force_path_style: bool,
}

pub fn resolve_settings(
    bucket_name: &str,
    region_env: &str,
    endpoint_env: &str,
    force_path_style: bool,
    floci_enabled: bool,
) -> Result<Settings, StorageError> {
    if bucket_name.is_empty() && !floci_enabled {
        return Err(StorageError::NotConfigured);
    }
    let (endpoint, region, bucket, force_path_style) = if floci_enabled {
        (
            if endpoint_env.is_empty() {
                DEFAULT_FLOCI_ENDPOINT.to_string()
            } else {
                endpoint_env.to_string()
            },
            if region_env.is_empty() {
                DEFAULT_FLOCI_REGION.to_string()
            } else {
                region_env.to_string()
            },
            if bucket_name.is_empty() {
                DEFAULT_FLOCI_BUCKET.to_string()
            } else {
                bucket_name.to_string()
            },
            // Local MinIO/Floci always uses path-style addressing.
            true,
        )
    } else {
        let endpoint = endpoint_env.to_string();
        let region = if region_env.is_empty() {
            "eu-west-1".to_string()
        } else {
            region_env.to_string()
        };
        // A custom S3-compatible endpoint almost always needs path-style
        // addressing (Go `resolveS3SettingsFromEnv`).
        let force_path_style = force_path_style || !endpoint.is_empty();
        (endpoint, region, bucket_name.to_string(), force_path_style)
    };
    Ok(Settings {
        endpoint,
        region,
        bucket,
        force_path_style,
    })
}

pub struct S3Client {
    client: aws_sdk_s3::Client,
    bucket: String,
}

impl S3Client {
    pub async fn new(settings: &Settings) -> Result<Self, StorageError> {
        let config = aws_config::defaults(aws_config::BehaviorVersion::latest())
            .region(aws_sdk_s3::config::Region::new(settings.region.clone()))
            .load()
            .await;
        let mut s3_options = aws_sdk_s3::config::Builder::from(&config);
        if !settings.endpoint.is_empty() {
            s3_options = s3_options.endpoint_url(settings.endpoint.clone());
        }
        if settings.force_path_style {
            s3_options = s3_options.force_path_style(true);
        }
        Ok(Self {
            client: aws_sdk_s3::Client::from_conf(s3_options.build()),
            bucket: settings.bucket.clone(),
        })
    }

    pub async fn put_object(
        &self,
        key: &str,
        content_type: &str,
        body: Vec<u8>,
    ) -> Result<(), StorageError> {
        self.client
            .put_object()
            .bucket(&self.bucket)
            .key(key)
            .content_type(content_type)
            .body(aws_sdk_s3::primitives::ByteStream::from(body))
            .send()
            .await
            .map_err(|error| StorageError::S3(error.to_string()))?;
        Ok(())
    }

    pub async fn get_object(&self, key: &str) -> Result<Vec<u8>, StorageError> {
        let output = self
            .client
            .get_object()
            .bucket(&self.bucket)
            .key(key)
            .send()
            .await
            .map_err(|error| StorageError::S3(error.to_string()))?;
        let body = output
            .body
            .collect()
            .await
            .map_err(|error| StorageError::S3(error.to_string()))?;
        Ok(body.into_bytes().to_vec())
    }

    pub async fn delete_object(&self, key: &str) -> Result<(), StorageError> {
        self.client
            .delete_object()
            .bucket(&self.bucket)
            .key(key)
            .send()
            .await
            .map_err(|error| StorageError::S3(error.to_string()))?;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::{resolve_settings, StorageError};

    #[test]
    fn settings_resolve() {
        assert_eq!(
            resolve_settings("", "eu-west-1", "", false, false),
            Err(StorageError::NotConfigured)
        );
        let production = resolve_settings("prod-bucket", "eu-west-1", "", false, false).unwrap();
        assert_eq!(production.bucket, "prod-bucket");
        let floci = resolve_settings("", "", "", false, true).unwrap();
        assert_eq!(floci.endpoint, "http://localhost:4566");
        assert_eq!(floci.bucket, "onion-dev");
    }
}
