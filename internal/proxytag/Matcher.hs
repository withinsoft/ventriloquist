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

main :: IO ()
main = do
  input <- ByteString.Lazy.getContents
  let response =
        case Aeson.decode input of
          Nothing ->
            Response
              { responseError = Just "error: could not decode message json"
              , responseMatch = Nothing
              }
          Just msg ->
            Response {responseError = Nothing, responseMatch = messageMatch msg}
      output = Aeson.encode response
  ByteString.Lazy.putStr output

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

data Response = Response
  { responseError :: Maybe Text
  , responseMatch :: Maybe Match
  } deriving (Eq, Read, Show, Generic)

instance Aeson.ToJSON Response where
  toEncoding = Aeson.genericToEncoding Aeson.defaultOptions

instance Aeson.FromJSON Response

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
  case Text.splitOn "\\" msg of
    [prefix, _] ->
      Just
        Matcher
          { matcherPrefix = Text.strip prefix
          , matcherSuffix = Nothing
          , matcherSystemMate = ""
          }
    _ -> Nothing
  

detectSigilsMatcher :: Text -> Maybe Matcher
detectSigilsMatcher = detectGenericMatcher

detectGenericMatcher :: Text -> Maybe Matcher
detectGenericMatcher msg =
  case Text.splitOn "text" msg of
    [prefix, _] ->
      Just
        Matcher
          { matcherPrefix = Text.strip prefix
          , matcherSuffix = Nothing
          , matcherSystemMate = ""
          }
    [prefix, _, suffix] ->
      Just
        Matcher
          { matcherPrefix = Text.strip prefix
          , matcherSuffix = Just (Text.strip suffix)
          , matcherSystemMate = ""
          }
    _ -> Nothing

detectMatcher :: Text -> Maybe Matcher
detectMatcher msg =
  getAlt
    (foldMap
       (\f -> Alt (f msg))
       [detectNamePrefixMatcher, detectSigilsMatcher, detectGenericMatcher])
