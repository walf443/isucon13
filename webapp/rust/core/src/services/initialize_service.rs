use crate::commands::initialize_command::{HaveInitializeCommand, InitializeCommand};
use crate::commands::CommandOutput;
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait InitializeService {
    async fn execute_command(&self) -> ServiceResult<CommandOutput>;
}

pub trait HaveInitializeService {
    type Service: InitializeService;

    fn initialize_service(&self) -> &Self::Service;
}

pub trait InitializeServiceImpl: Sync + HaveInitializeCommand {}

#[async_trait]
impl<T: InitializeServiceImpl> InitializeService for T {
    async fn execute_command(&self) -> ServiceResult<CommandOutput> {
        let output = self.initialize_command().execute().await?;
        Ok(output)
    }
}
