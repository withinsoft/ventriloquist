{-# LANGUAGE DeriveGeneric #-}
{-# LANGUAGE OverloadedStrings #-}
module Main where

import Data.Text (Text)
import qualified Data.Text as Text
import qualified Data.Aeson as Aeson
import Data.Vector (Vector)
import GHC.Generics
import qualified Data.ByteString.Lazy as ByteString.Lazy
import Data.Monoid
import Data.Char

main :: IO ()
main = do
  input <- ByteString.Lazy.getContents
  let response =
        either (\x -> ResponseError (Text.append "error: " x)) id $ do
          req <-
            maybe
              (Left "could not decode message json")
              Right
              (Aeson.decode input)
          case req of
            RequestDetectMatcher detectReq -> do
              found <-
                maybe
                  (Left "no matcher detected")
                  Right
                  (detectMatcher detectReq)
              return (ResponseMatcher found)
            RequestMatchMessage msg -> do
              found <- maybe (Left "no match found") Right (messageMatch msg)
              return (ResponseMatch found)
      output = Aeson.encode response
  ByteString.Lazy.putStr output

data Request
  = RequestMatchMessage Message
  | RequestDetectMatcher DetectMatcher
  deriving (Eq, Read, Show, Generic)

instance Aeson.FromJSON Request

instance Aeson.ToJSON Request where
  toEncoding = Aeson.genericToEncoding Aeson.defaultOptions

data Response
  = ResponseError Text
  | ResponseMatch Match
  | ResponseMatcher Matcher
  deriving (Eq, Read, Show, Generic)

instance Aeson.ToJSON Response where
  toEncoding = Aeson.genericToEncoding Aeson.defaultOptions

instance Aeson.FromJSON Response

data DetectMatcher = DetectMatcher
  { detectMatcherMessage :: Text
  , detectMatcherSystemMate :: Text
  } deriving (Eq, Read, Show, Generic)

instance Aeson.ToJSON DetectMatcher where
  toEncoding = Aeson.genericToEncoding Aeson.defaultOptions

instance Aeson.FromJSON DetectMatcher

data Message = Message
  { messageBody :: Text
  , messageMatchers :: Vector Matcher
  } deriving (Eq, Read, Show, Generic)

instance Aeson.ToJSON Message where
  toEncoding = Aeson.genericToEncoding Aeson.defaultOptions

instance Aeson.FromJSON Message

data Matcher = Matcher
  { matcherPrefix :: Text
  , matcherSuffix :: Maybe Text
  , matcherSystemMate :: Text
  } deriving (Eq, Read, Show, Generic)

instance Aeson.ToJSON Matcher where
  toEncoding = Aeson.genericToEncoding Aeson.defaultOptions

instance Aeson.FromJSON Matcher

data Match = Match
  { matchPrefix :: Text
  , matchSuffix :: Maybe Text
  , matchBody :: Text
  , matchSystemMate :: Text
  } deriving (Eq, Read, Show, Generic)

instance Aeson.ToJSON Match where
  toEncoding = Aeson.genericToEncoding Aeson.defaultOptions

instance Aeson.FromJSON Match


matchRaw :: Matcher -> Text -> Maybe Match
matchRaw matcher body = do
  withoutPrefix <- Text.stripPrefix (matcherPrefix matcher) body
  let withoutSuffix = do
        suffix <- matcherSuffix matcher
        Text.stripSuffix suffix withoutPrefix
  case withoutSuffix of
    Nothing ->
      Just
        Match
          { matchPrefix = matcherPrefix matcher
          , matchSuffix = Nothing
          , matchBody = withoutPrefix
          , matchSystemMate = matcherSystemMate matcher
          }
    Just b ->
      Just
        Match
          { matchPrefix = matcherPrefix matcher
          , matchSuffix = matcherSuffix matcher
          , matchBody = b
          , matchSystemMate = matcherSystemMate matcher
          }

-- This is just matchRaw but it runs Text.strip on the input and output
match :: Matcher -> Text -> Maybe Match
match matcher body = do
  result <- matchRaw matcher (Text.strip body)
  Just (result {matchBody = Text.strip (matchBody result)})

messageMatch :: Message -> Maybe Match
messageMatch message =
  getAlt
    (foldMap
       (\matcher -> Alt (match matcher (messageBody message)))
       (messageMatchers message))

detectNamePrefixMatcher :: Text -> Maybe Matcher
detectNamePrefixMatcher msg =
  getAlt (foldMap (Alt . matcherForIndicator) validPrefixIndicators)
  where
    validPrefixIndicators = ["\\", "/", ":", ">"]
    matcherForIndicator indicator =
      case Text.splitOn indicator msg of
        [prefix, _] ->
          Just
            Matcher
              { matcherPrefix = Text.strip prefix <> indicator
              , matcherSuffix = Nothing
              , matcherSystemMate = ""
              }
        _ -> Nothing
  

detectSigilsMatcher :: Text -> Maybe Matcher
detectSigilsMatcher msg
  | Text.length msg == 1 && not (isAlphaNum (Text.head msg)) =
    Just
      Matcher
        {matcherPrefix = msg, matcherSuffix = Nothing, matcherSystemMate = ""}
  | Text.length msg == 2 && not (Text.any isAlphaNum msg) =
    Just
      Matcher
        { matcherPrefix = Text.singleton (Text.head msg)
        , matcherSuffix = Just (Text.singleton (Text.last msg))
        , matcherSystemMate = ""
        }
  | otherwise = Nothing
  

detectGenericMatcher :: Text -> Maybe Matcher
detectGenericMatcher msg =
  case Text.splitOn "text" msg of
    [prefix, suffix] ->
      Just
        Matcher
          { matcherPrefix = Text.strip prefix
          , matcherSuffix = Just (Text.strip suffix)
          , matcherSystemMate = ""
          }
    _ -> Nothing

detectMatcher :: DetectMatcher -> Maybe Matcher
detectMatcher detectReq = do
  foundMatcher <-
    getAlt
      (foldMap
         (\f -> Alt (f msg))
         [detectGenericMatcher, detectNamePrefixMatcher, detectSigilsMatcher])
  return foundMatcher {matcherSystemMate = sysMate}
  where
    msg = detectMatcherMessage detectReq
    sysMate = detectMatcherSystemMate detectReq
