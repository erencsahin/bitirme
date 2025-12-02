use anyhow::Result;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use tracing::{info, warn};

#[derive(Debug, Deserialize)]
struct ValidateTokenResponse {
    status: String,
    data: Option<TokenData>,
}

#[derive(Debug, Deserialize)]
struct TokenData {
    valid: bool,
    #[serde(rename = "userId")]
    user_id: String,
    email: String,
}

pub struct UserServiceClient {
    base_url: String,
    client: Client,
}

impl UserServiceClient {
    pub fn new(base_url: String) -> Self {
        Self {
            base_url,
            client: Client::new(),
        }
    }

    pub async fn validate_token(&self, token: &str) -> Result<bool> {
        let url = format!("{}/api/auth/validate", self.base_url);

        let response = self
            .client
            .post(&url)
            .header("Authorization", format!("Bearer {}", token))
            .send()
            .await?;

        if response.status().is_success() {
            let result: ValidateTokenResponse = response.json().await?;
            let is_valid = result.status == "success" 
                && result.data.as_ref().map(|d| d.valid).unwrap_or(false);
            
            info!("Token validation result: {}", is_valid);
            Ok(is_valid)
        } else {
            warn!("Token validation failed: {}", response.status());
            Ok(false)
        }
    }

    pub async fn get_user_id_from_token(&self, token: &str) -> Result<Option<String>> {
        let url = format!("{}/api/auth/validate", self.base_url);

        let response = self
            .client
            .post(&url)
            .header("Authorization", format!("Bearer {}", token))
            .send()
            .await?;

        if response.status().is_success() {
            let result: ValidateTokenResponse = response.json().await?;
            Ok(result.data.map(|d| d.user_id))
        } else {
            Ok(None)
        }
    }
}