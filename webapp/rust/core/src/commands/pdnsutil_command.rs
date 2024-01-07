use crate::commands::{CommandOutput, CommandResult};
use async_trait::async_trait;

#[cfg_attr(any(feature = "test", test), mockall::automock)]
#[async_trait]
pub trait PDNSUtilCommand {
    async fn add_record(&self, name: &str, domain: &str) -> CommandResult<CommandOutput>;
}

pub trait HavePDNSUtilCommand {
    type Command: PDNSUtilCommand;

    fn pdnsutil_command(&self) -> &Self::Command;
}
