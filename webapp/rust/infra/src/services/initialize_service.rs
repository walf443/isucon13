use crate::commands::initialize_command::InitializeCommandInfra;
use isupipe_core::commands::initialize_command::HaveInitializeCommand;
use isupipe_core::services::initialize_service::InitializeServiceImpl;

#[derive(Clone)]
pub struct InitializeServiceInfra {
    initialize_command: InitializeCommandInfra,
}

impl InitializeServiceInfra {
    pub fn new() -> Self {
        Self {
            initialize_command: InitializeCommandInfra {},
        }
    }
}

impl HaveInitializeCommand for InitializeServiceInfra {
    type Command = InitializeCommandInfra;

    fn initialize_command(&self) -> &Self::Command {
        &self.initialize_command
    }
}

impl InitializeServiceImpl for InitializeServiceInfra {}
