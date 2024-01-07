use crate::commands::pdnsutil_command::HavePDNSUtilCommand;
use crate::db::HaveDBPool;
use crate::repos::livestream_comment_report_repository::HaveLivestreamCommentReportRepository;
use crate::repos::livestream_comment_repository::HaveLivestreamCommentRepository;
use crate::repos::livestream_repository::HaveLivestreamRepository;
use crate::repos::livestream_tag_repository::HaveLivestreamTagRepository;
use crate::repos::livestream_viewers_history_repository::HaveLivestreamViewersHistoryRepository;
use crate::repos::ng_word_repository::HaveNgWordRepository;
use crate::repos::reaction_repository::HaveReactionRepository;
use crate::repos::reservation_slot_repository::HaveReservationSlotRepository;
use crate::repos::tag_repository::HaveTagRepository;
use crate::repos::theme_repository::HaveThemeRepository;
use crate::repos::user_repository::HaveUserRepository;

pub trait RepositoryManager:
    Sync
    + HaveDBPool
    + HaveLivestreamCommentReportRepository
    + HaveLivestreamCommentRepository
    + HaveLivestreamRepository
    + HaveLivestreamTagRepository
    + HaveLivestreamViewersHistoryRepository
    + HaveNgWordRepository
    + HaveReactionRepository
    + HaveReservationSlotRepository
    + HaveTagRepository
    + HaveThemeRepository
    + HaveUserRepository
    + HavePDNSUtilCommand
{
}

#[cfg(test)]
pub mod tests {
    use crate::commands::pdnsutil_command::{HavePDNSUtilCommand, MockPDNSUtilCommand};
    use crate::db::{DBPool, HaveDBPool};
    use crate::repos::livestream_comment_report_repository::{
        HaveLivestreamCommentReportRepository, MockLivestreamCommentReportRepository,
    };
    use crate::repos::livestream_comment_repository::{
        HaveLivestreamCommentRepository, MockLivestreamCommentRepository,
    };
    use crate::repos::livestream_repository::{HaveLivestreamRepository, MockLivestreamRepository};
    use crate::repos::livestream_tag_repository::{
        HaveLivestreamTagRepository, MockLivestreamTagRepository,
    };
    use crate::repos::livestream_viewers_history_repository::{
        HaveLivestreamViewersHistoryRepository, MockLivestreamViewersHistoryRepository,
    };
    use crate::repos::manager::RepositoryManager;
    use crate::repos::ng_word_repository::{HaveNgWordRepository, MockNgWordRepository};
    use crate::repos::reaction_repository::{HaveReactionRepository, MockReactionRepository};
    use crate::repos::reservation_slot_repository::{
        HaveReservationSlotRepository, MockReservationSlotRepository,
    };
    use crate::repos::tag_repository::{HaveTagRepository, MockTagRepository};
    use crate::repos::theme_repository::{HaveThemeRepository, MockThemeRepository};
    use crate::repos::user_repository::{HaveUserRepository, MockUserRepository};
    use crate::services::user_service::UserServiceImpl;

    pub struct MockRepositoryManager {
        db_pool: DBPool,
        pub mock_livestream_comment_report_repo: MockLivestreamCommentReportRepository,
        pub mock_livestream_comment_repo: MockLivestreamCommentRepository,
        pub mock_livestream_repo: MockLivestreamRepository,
        pub mock_livestream_tag_repo: MockLivestreamTagRepository,
        pub mock_livestream_viewers_history_repo: MockLivestreamViewersHistoryRepository,
        pub mock_ng_word_repo: MockNgWordRepository,
        pub mock_reaction_repo: MockReactionRepository,
        pub mock_reservation_slot_repo: MockReservationSlotRepository,
        pub mock_tag_repo: MockTagRepository,
        pub mock_theme_repo: MockThemeRepository,
        pub mock_user_repo: MockUserRepository,
        pub mock_pdns_util_command: MockPDNSUtilCommand,
    }

    impl MockRepositoryManager {
        pub fn new(db_pool: DBPool) -> Self {
            Self {
                db_pool,
                mock_livestream_comment_report_repo: Default::default(),
                mock_livestream_comment_repo: Default::default(),
                mock_livestream_repo: Default::default(),
                mock_livestream_tag_repo: Default::default(),
                mock_livestream_viewers_history_repo: Default::default(),
                mock_ng_word_repo: Default::default(),
                mock_reaction_repo: Default::default(),
                mock_reservation_slot_repo: Default::default(),
                mock_tag_repo: Default::default(),
                mock_theme_repo: Default::default(),
                mock_user_repo: Default::default(),
                mock_pdns_util_command: Default::default(),
            }
        }
    }

    impl HaveDBPool for MockRepositoryManager {
        fn get_db_pool(&self) -> &DBPool {
            &self.db_pool
        }
    }

    impl HaveLivestreamCommentReportRepository for MockRepositoryManager {
        type Repo = MockLivestreamCommentReportRepository;

        fn livestream_comment_report_repo(&self) -> &Self::Repo {
            &self.mock_livestream_comment_report_repo
        }
    }

    impl HaveLivestreamCommentRepository for MockRepositoryManager {
        type Repo = MockLivestreamCommentRepository;

        fn livestream_comment_repo(&self) -> &Self::Repo {
            &self.mock_livestream_comment_repo
        }
    }

    impl HaveLivestreamRepository for MockRepositoryManager {
        type Repo = MockLivestreamRepository;

        fn livestream_repo(&self) -> &Self::Repo {
            &self.mock_livestream_repo
        }
    }

    impl HaveLivestreamTagRepository for MockRepositoryManager {
        type Repo = MockLivestreamTagRepository;

        fn livestream_tag_repo(&self) -> &Self::Repo {
            &self.mock_livestream_tag_repo
        }
    }

    impl HaveLivestreamViewersHistoryRepository for MockRepositoryManager {
        type Repo = MockLivestreamViewersHistoryRepository;

        fn livestream_viewers_history_repo(&self) -> &Self::Repo {
            &self.mock_livestream_viewers_history_repo
        }
    }

    impl HaveNgWordRepository for MockRepositoryManager {
        type Repo = MockNgWordRepository;

        fn ng_word_repo(&self) -> &Self::Repo {
            &self.mock_ng_word_repo
        }
    }

    impl HaveReactionRepository for MockRepositoryManager {
        type Repo = MockReactionRepository;

        fn reaction_repo(&self) -> &Self::Repo {
            &self.mock_reaction_repo
        }
    }

    impl HaveReservationSlotRepository for MockRepositoryManager {
        type Repo = MockReservationSlotRepository;

        fn reservation_slot_repo(&self) -> &Self::Repo {
            &self.mock_reservation_slot_repo
        }
    }

    impl HaveTagRepository for MockRepositoryManager {
        type Repo = MockTagRepository;

        fn tag_repo(&self) -> &Self::Repo {
            &self.mock_tag_repo
        }
    }

    impl HaveThemeRepository for MockRepositoryManager {
        type Repo = MockThemeRepository;

        fn theme_repo(&self) -> &Self::Repo {
            &self.mock_theme_repo
        }
    }

    impl HaveUserRepository for MockRepositoryManager {
        type Repo = MockUserRepository;

        fn user_repo(&self) -> &Self::Repo {
            &self.mock_user_repo
        }
    }

    impl HavePDNSUtilCommand for MockRepositoryManager {
        type Command = MockPDNSUtilCommand;

        fn pdnsutil_command(&self) -> &Self::Command {
            &self.mock_pdns_util_command
        }
    }

    impl RepositoryManager for MockRepositoryManager {}
    impl UserServiceImpl for MockRepositoryManager {}
}
